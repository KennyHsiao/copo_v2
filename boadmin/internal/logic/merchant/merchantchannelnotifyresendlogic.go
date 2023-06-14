package merchant

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantChannelNotifyResendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantChannelNotifyResendLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantChannelNotifyResendLogic {
	return MerchantChannelNotifyResendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantChannelNotifyResendLogic) MerchantChannelNotifyResend(req *types.MerhcantChannelNotifyResendRequest) error {
	var channelChangeNotify types.ChannelChangeNotifyX
	if err := l.svcCtx.MyDB.Table("mc_channel_change_notify").
		Where("merchant_code = ?", req.MerchantCode).Take(&channelChangeNotify).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	var jsonSlice []types.MerchantChannelPayTypeLimit
	jsonArrStr := []byte(channelChangeNotify.LastNotifyMessage)
	json.Unmarshal(jsonArrStr, &jsonSlice)

	currencyCode := jsonSlice[0].Currency
	merchantOpenChannel := types.MerchantOpenChannel{
		MerchantId:                   req.MerchantCode,
		CurrencyCode:                 currencyCode,
		MerchantChannelPayTypeLimits: jsonSlice,
	}

	if isOk := merchantsService.DoCallNoticeURL(l.ctx, merchantOpenChannel, channelChangeNotify.NotifyUrl); isOk {
		logx.WithContext(l.ctx).Infof("通知商户渠道变动: 商户:%s 通知成功", channelChangeNotify.MerchantCode)
		channelChangeNotify.Status = "1"
		err := l.svcCtx.MyDB.Table("mc_channel_change_notify").Where("id = ?", channelChangeNotify.ID).Update("status", channelChangeNotify.Status).Error
		if err != nil {
			logx.WithContext(l.ctx).Errorf("更新通知成功状态,失败, merchantCode : %+v", channelChangeNotify.MerchantCode)
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	} else {
		channelChangeNotify.Status = "0"
		err := l.svcCtx.MyDB.Table("mc_channel_change_notify").Where("id = ?", channelChangeNotify.ID).Update("status", channelChangeNotify.Status).Error
		if err != nil {
			logx.WithContext(l.ctx).Errorf("更新通知失败状态,失败, merchantCode : %+v", channelChangeNotify.MerchantCode)
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	return nil
}
