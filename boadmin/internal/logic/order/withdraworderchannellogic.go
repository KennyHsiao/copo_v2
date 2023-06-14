package order

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderChannelLogic {
	return WithdrawOrderChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawOrderChannelLogic) WithdrawOrderChannel(req types.WithdrawOrderChannelRequest) (resp *types.WithdrawOrderChannelResponse, err error) {
	var channelCodes []types.ChannelCodeAndHandlingFee
	var terms []string
	var terms2 []string
	for _, s := range req.Status {
		terms = append(terms, fmt.Sprintf("chn.status = '%s'", s)) // 狀態(关闭=0;开启=1;维护=2) 2022-08-03 客服说不要只捞开启的渠道
	}
	for _, p := range req.IsProxy {
		terms2 = append(terms2, fmt.Sprintf("chn.is_proxy = '%s'", p)) // 狀態(关闭=0;开启=1;维护=2) 2022-08-03 客服说不要只捞开启的渠道
	}

	if req.IsXf == "1" {

		term := strings.Join(terms, " OR ")
		term2 := strings.Join(terms2, " OR ")
		if err = l.svcCtx.MyDB.Select("chn.code, chn.channel_withdraw_charge, chn.name").Table("ch_channels chn").
			Where("chn.currency_code = ?", req.CurrencyCode).Where(term).Where(term2).Find(&channelCodes).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		resp = &types.WithdrawOrderChannelResponse{
			List: channelCodes,
		}

	} else {
		term := strings.Join(terms, " OR ")
		term2 := strings.Join(terms2, " OR ")

		if err = l.svcCtx.MyDB.Select("chn.code, chn.channel_withdraw_charge, chn.name, ccpt.fee, ccpt.handling_fee, ccpt.is_rate").
			Table("ch_channel_pay_types AS ccpt").
			Joins("LEFT JOIN ch_channels AS chn ON ccpt.channel_code = chn.`code`").
			Where("chn.currency_code = ? AND ccpt.`status` = ? AND ccpt.pay_type_code = ?", req.CurrencyCode, "1", "DF").
			Where(term).Where(term2).Find(&channelCodes).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		resp = &types.WithdrawOrderChannelResponse{
			List: channelCodes,
		}
	}

	return
}
