package channeldata

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelDataQueryForBKProxyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataQueryForBKProxyLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataQueryForBKProxyLogic {
	return ChannelDataQueryForBKProxyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataQueryForBKProxyLogic) ChannelDataQueryForBKProxy(req *types.ChannelDataQueryForBKProxyRequestX) (resp *types.ChannelDataQueryForBKProxyResponse, err error) {

	var channels []types.ChannelData
	db := l.svcCtx.MyDB

	if len(req.Status) > 0 {
		for _, status := range req.Status {
			db = db.Where("cc.status = ?", status)
		}
	}

	selectX := "cc.id," +
		"cc.code," +
		"cc.name," +
		"cc.project_name," +
		"cc.is_proxy," +
		"cc.is_nz_pre," +
		"cc.api_url," +
		"cc.currency_code," +
		"cc.channel_withdraw_charge," +
		"cc.balance," +
		"cc.balance_limit," +
		"cc.status," +
		"cc.device," +
		"cc.mer_id," +
		"cc.mer_key," +
		"cc.pay_url," +
		"cc.pay_query_url," +
		"cc.pay_query_balance_url," +
		"cc.proxy_pay_url," +
		"cc.proxy_pay_query_url," +
		"cc.proxy_pay_query_balance_url," +
		"cc.white_list," +
		"cc.pay_type_map," +
		"cc.channel_port," +
		"cc.withdraw_balance," +
		"cc.proxypay_balance"

	if err := db.Table("ch_channels cc").
		Joins("LEFT JOIN ch_channel_pay_types ccpt ON cc.code = ccpt.channel_code").
		Scopes(gormx.Sort(req.Orders)).
		Select(selectX).
		Where("ccpt.pay_type_code = ? ", "DF").
		Where("cc.currency_code = ?", req.CurrencyCode).
		Find(&channels).Error; err != nil {
	}

	return &types.ChannelDataQueryForBKProxyResponse{
		List: channels,
	}, nil
}
