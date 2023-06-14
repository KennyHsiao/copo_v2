package channelbalance

import (
	"com.copo/bo_service/boadmin/internal/service/callNoticeUrlService"
	channelBalanceBalance "com.copo/bo_service/boadmin/internal/service/channelBalanceService"
	"com.copo/bo_service/boadmin/internal/service/channelDataService"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"fmt"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBalanceScheduleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBalanceScheduleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBalanceScheduleUpdateLogic {
	return ChannelBalanceScheduleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBalanceScheduleUpdateLogic) ChannelBalanceScheduleUpdate() error {
	var channelDataList []types.ChannelDataUpdate

	if err := l.svcCtx.MyDB.Table("ch_channels").Where("status = '1'").Find(&channelDataList).Error; err != nil {

	}

	channelBalanceNotify, _ := channelBalanceBalance.UpdateChannelBalanceForSchedule(l.ctx, l.svcCtx, channelDataList)

	if len(channelBalanceNotify.FailList) > 0 {
		err := callNoticeUrlService.CallNoticeUrlForChannelBalanceFail(l.ctx, l.svcCtx, channelBalanceNotify.FailList)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("更新馀额错误通知失敗:%s", err.Error())
		}
	}

	if len(channelBalanceNotify.SuccessList) > 0 {
		newChannelDataList, err := channelDataService.Query_N_ChannelData(l.svcCtx.MyDB, channelBalanceNotify.SuccessList)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("更新馀额通知失敗:%s", err.Error())
		}

		var systemParams types.SystemParams
		if errs := l.svcCtx.MyDB.Table("bs_system_params").
			Where("name = 'channelBalanceLimit'").Find(&systemParams).Error; errs != nil {
			logx.WithContext(l.ctx).Errorf("查询系统设定值失敗:%s", errs.Error())
		}

		systemLimit, errf := strconv.ParseFloat(systemParams.Value, 64)
		if errf != nil {
			logx.WithContext(l.ctx).Errorf("系统设定值转换失敗:%s", errf.Error())
		}

		for _, data := range *newChannelDataList {
			if data.BalanceLimit > 0 {
				if data.WithdrawBalance > data.BalanceLimit {
					notifyMsg := types.TelegramNotifyRequest{
						Message: fmt.Sprintf("渠道名称： %s(%s) \n目前余额： %f", data.Name, data.CurrencyCode, data.WithdrawBalance),
					}
					err := callNoticeUrlService.CallNoticeUrlForChannelBalanceSuccess(l.ctx, l.svcCtx, notifyMsg)
					if err != nil {
						logx.WithContext(l.ctx).Errorf("馀额报警通知失敗:%s", err.Error())
					}
				} else if data.ProxypayBalance > data.BalanceLimit {
					notifyMsg := types.TelegramNotifyRequest{
						Message: fmt.Sprintf("渠道名称： %s(%s) \n目前余额： %f", data.Name, data.CurrencyCode, data.ProxypayBalance),
					}
					err := callNoticeUrlService.CallNoticeUrlForChannelBalanceSuccess(l.ctx, l.svcCtx, notifyMsg)
					if err != nil {
						logx.WithContext(l.ctx).Errorf("馀额报警通知失敗:%s", err.Error())
					}
				}
			} else {
				if data.WithdrawBalance > systemLimit {
					notifyMsg := types.TelegramNotifyRequest{
						Message: fmt.Sprintf("渠道名称： %s(%s) \n目前余额： %f", data.Name, data.CurrencyCode, data.WithdrawBalance),
					}
					err := callNoticeUrlService.CallNoticeUrlForChannelBalanceSuccess(l.ctx, l.svcCtx, notifyMsg)
					if err != nil {
						logx.WithContext(l.ctx).Errorf("馀额报警通知失敗:%s", err.Error())
					}
				} else if data.ProxypayBalance > systemLimit {
					notifyMsg := types.TelegramNotifyRequest{
						Message: fmt.Sprintf("渠道名称： %s(%s) \n目前余额： %f", data.Name, data.CurrencyCode, data.ProxypayBalance),
					}
					err := callNoticeUrlService.CallNoticeUrlForChannelBalanceSuccess(l.ctx, l.svcCtx, notifyMsg)
					if err != nil {
						logx.WithContext(l.ctx).Errorf("馀额报警通知失敗:%s", err.Error())
					}
				}
			}
		}
	}

	return nil
}
