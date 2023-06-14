package channeldata

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ChannelDataUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataUpdateLogic {
	return ChannelDataUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataUpdateLogic) ChannelDataUpdate(req types.ChannelDataUpdateRequest) error {
	err := l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		channelData := &types.ChannelDataUpdate{
			ChannelDataUpdateRequest: req,
		}

		//查詢是否有該渠道
		isExist := model.NewChannel(db).CheckChannelIsExist(req.Code)
		if !isExist {
			return errorz.New(response.CHANNEL_IS_NOT_EXIST)
		}

		//检查请求参数
		paramErr := model.CheckRequestUpdateValue(req)
		if paramErr != nil {
			return paramErr
		}

		if len(req.ChannelPayTypeList) > 0 {
			var updateChannelPayTypeList []types.ChannelPayTypeUpdateRequest
			var createChannelPayTypeList []types.ChannelPayTypeCreateRequest
			datas := make(map[string]string)
			for _, m := range req.ChannelPayTypeList {
				//datas[m.PayTypeCode] = m.MapCode
				if m.ID > 0 { //更新
					//同步新增到channel_pay_type
					updateChannelPayType := types.ChannelPayTypeUpdateRequest{
						ID:                m.ID,
						Code:              m.Code, //渠道號+支付代碼
						ChannelCode:       m.ChannelCode,
						PayTypeCode:       m.PayTypeCode,
						Status:            m.Status,
						MapCode:           m.MapCode,
						Fee:               m.Fee,
						HandlingFee:       m.HandlingFee,
						MaxInternalCharge: m.MaxInternalCharge,
						DailyTxLimit:      m.DailyTxLimit,
						SingleMaxCharge:   m.SingleMaxCharge,
						SingleMinCharge:   m.SingleMinCharge,
						FixedAmount:       m.FixedAmount,
						BillDate:          m.BillDate,
						IsProxy:           m.IsProxy,
						IsRate:            m.IsRate,
						Device:            m.Device,
					}
					model.NewChannelPayType(db).SingleUpdateChannelPayType(updateChannelPayType)
					//updateChannelPayTypeList = append(updateChannelPayTypeList, updateChannelPayType)
				} else { //新增
					createChannelPayType := types.ChannelPayTypeCreateRequest{
						Code:              m.ChannelCode + m.PayTypeCode, //渠道號+支付代碼
						ChannelCode:       m.ChannelCode,
						PayTypeCode:       m.PayTypeCode,
						Status:            constants.OFF,
						MapCode:           m.MapCode,
						Fee:               m.Fee,
						HandlingFee:       m.HandlingFee,
						MaxInternalCharge: m.MaxInternalCharge,
						DailyTxLimit:      m.DailyTxLimit,
						SingleMaxCharge:   m.SingleMaxCharge,
						SingleMinCharge:   m.SingleMinCharge,
						FixedAmount:       m.FixedAmount,
						BillDate:          m.BillDate,
						IsProxy:           m.IsProxy,
						IsRate:            "0",
						Device:            "All",
					}
					createChannelPayTypeList = append(createChannelPayTypeList, createChannelPayType)
				}
			}
			if len(updateChannelPayTypeList) > 0 {
				updateErr := model.NewChannelPayType(db).UpdateChannelPayType(updateChannelPayTypeList)
				if updateErr != nil {
					return updateErr
				}
			}
			if len(createChannelPayTypeList) > 0 {
				createErr := model.NewChannelPayType(db).InsertChannelPayType(createChannelPayTypeList)
				if createErr != nil {
					return createErr
				}
			}
			p, _ := json.Marshal(datas)
			channelData.PayTypeMap = string(p)
			logx.Info("channelData.PayTypeMap: ", channelData.PayTypeMap)
		}
		//刪除舊資料
		errDelete := db.Table("ch_channel_banks").Where("channel_code = ?", req.Code).Delete(&types.ChannelBank{}).Error
		if errDelete != nil {
			return errDelete
		}
		if len(req.BankCodeMapList) > 0 {
			var createChannelBankList []types.ChannelBankCreateRequest
			for _, m := range req.BankCodeMapList {
				createChannelBank := types.ChannelBankCreateRequest{
					ChannelCode: channelData.Code,
					BankNo:      m.BankNo,
					MapCode:     m.MapCode,
				}
				createChannelBankList = append(createChannelBankList, createChannelBank)
			}
			if len(createChannelBankList) > 0 {
				if createErr := model.NewChannelBank(db).InsertChannelBank(createChannelBankList); createErr != nil {
					return createErr
				}
			}
		}
		l.Info("更新渠道資料:", channelData)
		updateChannel := &types.ChannelDataUpdate{
			ChannelDataUpdateRequest: req,
		}

		//更新渠道
		err := db.Table("ch_channels").Save(updateChannel).Error

		return err
	})

	// 发送渠道修改通知
	merchantsService.ChannelChangeNotify(l.svcCtx.MyDB, l.ctx, l.svcCtx, req.CurrencyCode)

	if err != nil {
		return err
	}

	return nil
}
