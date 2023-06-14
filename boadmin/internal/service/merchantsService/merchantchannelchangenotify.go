package merchantsService

import (
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"time"
)

func ChannelChangeNotify(db *gorm.DB, ctx context.Context, scvCtx *svc.ServiceContext, currencyCode string) (resErr error) {
	logx.WithContext(ctx).Infof("发送最新可用通道通知")

	//取得渠道變更通知商戶名單
	var channelChangeNotifys []types.ChannelChangeNotifyX
	if err := db.Table("mc_channel_change_notify").
		Where("is_channel_change_notify = '1'").Find(&channelChangeNotifys).Error; err != nil {
		logx.WithContext(ctx).Errorf("查询渠道变更通知商户名单错误")
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for _, notify := range channelChangeNotifys {
		tx := db.Begin()
		if len(notify.NotifyUrl) > 0 {
			var merchantChannelRates []*types.MerchantOrderRateListViewX
			merchantChannelRates, err := ordersService.GetAllMerchantChannelRate(db, notify.MerchantCode, currencyCode)
			if err != nil {
				notify.Status = "0"
				err2 := tx.Table("mc_channel_change_notify").Where("id = ?", notify.ID).Update("status", notify.Status).Error
				if err2 != nil {
					tx.Rollback()
					return errorz.New(response.DATABASE_FAILURE, err2.Error())
				}
				return err
			}

			//if len(merchantChannelRates) == 0 {
			//	logx.WithContext(ctx).Infof("商户{}:没有可用通道,不发送通知", notify.MerchantCode)
			//
			//}

			var saveMessage string
			var merchantChannelPayTypeLimits []types.MerchantChannelPayTypeLimit
			if len(merchantChannelRates) > 0 {
				for _, rate := range merchantChannelRates {
					merchantChannelPayTypeLimit := types.MerchantChannelPayTypeLimit{
						PayType:            rate.PayTypeCode,
						PayTypeName:        rate.PayTypeName,
						SingleLimitMaxmum:  rate.SingleMaxCharge,
						SingleLimitMinimum: rate.SingleMinCharge,
						PayTypeImageUrl:    scvCtx.Config.ResourceHost + rate.PayTypeImgUrl,
						Fixed:              rate.FixedAmount,
						Device:             rate.Device,
						Currency:           currencyCode,
					}
					merchantChannelPayTypeLimits = append(merchantChannelPayTypeLimits, merchantChannelPayTypeLimit)
				}
			} else {
				merchantChannelPayTypeLimit := types.MerchantChannelPayTypeLimit{
					Currency: currencyCode,
				}
				merchantChannelPayTypeLimits = append(merchantChannelPayTypeLimits, merchantChannelPayTypeLimit)
			}

			b, errM := json.Marshal(merchantChannelPayTypeLimits)
			if errM != nil {
				tx.Rollback()
				return errM
			}
			saveMessage = string(b)
			logx.WithContext(ctx).Infof("要纪录起来的发送资讯: %v", saveMessage)

			//比對發送內容是否有變更，有的話才發送
			lastNotifyMessage := notify.LastNotifyMessage
			logx.WithContext(ctx).Infof("上一次发送资讯: %v", lastNotifyMessage)
			if saveMessage != lastNotifyMessage {
				merchantOpenChannel := types.MerchantOpenChannel{
					MerchantId:                   notify.MerchantCode,
					CurrencyCode:                 currencyCode,
					MerchantChannelPayTypeLimits: merchantChannelPayTypeLimits,
				}

				//發送通知
				logx.WithContext(ctx).Infof("渠道资讯有变更，發送通知商户'%s'，发送资讯='%+v'", notify.MerchantCode, merchantOpenChannel)

				var minDelaySeconds int64 = 10
				if len(notify.NotifyUrl) > 0 {
					go func(notify2 types.ChannelChangeNotifyX) {
						var finalStatus bool
						for i := 0; i < 5; i++ {
							startTime := time.Now().Unix()
							logx.WithContext(ctx).Infof("通知商户渠道变动: 第%d次通知 商户:%s, NotifyUrl:%s, request: %+v", i+1, notify2.MerchantCode, notify2.NotifyUrl, merchantOpenChannel)
							if isOk := DoCallNoticeURL(ctx, merchantOpenChannel, notify2.NotifyUrl); isOk {
								logx.WithContext(ctx).Infof("通知商户渠道变动: 商户:%s 通知成功", notify2.MerchantCode)
								notify.Status = "1"
								err := db.Table("mc_channel_change_notify").Where("id = ?", notify2.ID).Update("status", notify2.Status).Error
								if err != nil {
									logx.WithContext(ctx).Errorf("更新通知成功状态,失败, merchantCode : %v", notify2.MerchantCode)
								}
								finalStatus = isOk
								break
							}
							endTime := time.Now().Unix()
							secondsDiff := endTime - startTime
							if secondsDiff < minDelaySeconds {
								sleepTime := time.Duration(minDelaySeconds-secondsDiff) * time.Second
								time.Sleep(sleepTime)
							}
						}
						if !finalStatus {
							notify2.Status = "0"
							err := db.Table("mc_channel_change_notify").Where("id = ?", notify2.ID).Update("status", notify2.Status).Error
							if err != nil {
								logx.WithContext(ctx).Errorf("更新通知失败状态,失败, merchantCode : %+v", notify2.MerchantCode)
							}
							msg := fmt.Sprintf("渠道变更通知商户失败，商户号： '%s'", notify2.MerchantCode)
							DoCallLineSendURL(ctx, scvCtx, msg)
						}
					}(notify)
				}

				//更新最後一次發送內容
				notify.LastNotifyMessage = saveMessage
			}
			//儲存最後一次發送內容 for 避免重複發送
			if err := tx.Table("mc_channel_change_notify").Updates(notify).Error; err != nil {
				tx.Rollback()
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			tx.Commit()
		}
	}

	return nil
}

func DoCallNoticeURL(ctx context.Context, merchantOpenChannel types.MerchantOpenChannel, notifyUrl string) (isOk bool) {
	span := trace.SpanFromContext(ctx)
	res, ChnErr := gozzle.Post(notifyUrl).Timeout(20).Trace(span).JSON(merchantOpenChannel)

	if ChnErr != nil {
		logx.WithContext(ctx).Errorf("通知商户渠道变动 gozzle Error: %s, Error:%s", merchantOpenChannel.MerchantId, ChnErr.Error())
		return false
	}
	resString := string(res.Body()[:])
	if res.Status() < 200 || res.Status() >= 300 {
		logx.WithContext(ctx).Errorf("通知商户渠道变动 状态码错误: %s, HttpStatus:%d, Response:%s",
			merchantOpenChannel.MerchantId, res.Status(), resString)
		return false
	}
	if resString == "success" {
		logx.WithContext(ctx).Infof("通知商户渠道变动 通知成功: %s, Response:%s", merchantOpenChannel.MerchantId, resString)
		return true
	} else {
		logx.WithContext(ctx).Errorf("通知商户渠道变动 商户回复错误: %s, Response:%s", merchantOpenChannel.MerchantId, resString)
		return false
	}
}

func DoCallLineSendURL(ctx context.Context, svcCtx *svc.ServiceContext, message string) error {
	span := trace.SpanFromContext(ctx)
	notifyUrl := fmt.Sprintf("%s:%d/line/send", svcCtx.Config.LineSend.Host, svcCtx.Config.LineSend.Port)
	data := struct {
		Message string `json:"message"`
	}{
		Message: message,
	}

	lineKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.LineKey, svcCtx.Config.ApiKey.PublicKey)
	if errk != nil {
		logx.WithContext(ctx).Errorf("MicroServiceEncrypt: %s", errk.Error())
		return errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	res, errx := gozzle.Post(notifyUrl).Timeout(20).Trace(span).Header("authenticationLineKey", lineKey).JSON(data)
	if res != nil {
		logx.WithContext(ctx).Info("response Status:", res.Status())
		logx.WithContext(ctx).Info("response Body:", string(res.Body()))
	}
	if errx != nil {
		logx.WithContext(ctx).Errorf("call Channel cha: %s", errx.Error())
		return errorz.New(response.GENERAL_EXCEPTION, errx.Error())
	} else if res.Status() != 200 {
		return errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call channelApp httpStatus:%d", res.Status()))
	}

	return nil
}
