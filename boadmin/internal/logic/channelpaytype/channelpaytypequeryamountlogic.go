package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelPayTypeQueryAmountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeQueryAmountLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeQueryAmountLogic {
	return ChannelPayTypeQueryAmountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeQueryAmountLogic) ChannelPayTypeQueryAmount(req *types.ChannelPayTypeQueryAmountRequestX) (resp *types.ChannelPayTypeQueryAmountResponse, err error) {
	var channelPayTypeQueryAmounts []types.ChannelPayTypeQueryAmount
	var count int64
	db := l.svcCtx.MyDB

	db = db.Where(" cpt.status != '0'")

	if len(req.ChannelCode) > 0 {
		db = db.Where(" cpt.channel_code = ?", req.ChannelCode)
	}
	if len(req.ChannelName) > 0 {
		db = db.Where(" c.name like ?", "%"+req.ChannelName+"%")
	}
	if len(req.PayTypeCode) > 0 {
		db = db.Where(" cpt.pay_type_code = ?", req.PayTypeCode)
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where(" c.`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.Status) > 0 {
		db = db.Where(" cpt.status = ?", req.Status)
	}
	if len(req.IsProxy) > 0 {
		db = db.Where(" c.`is_proxy` = ?", req.IsProxy)
	}

	selectX := "pt.name as pay_type_name, " +
		"c.is_proxy as is_proxy, " +
		"c.name as channel_name, " +
		"cpt.fee as fee, " +
		"cpt.handling_fee as handling_fee, " +
		"cpt.single_min_charge as single_min_charge, " +
		"cpt.single_max_charge as single_max_charge, " +
		"cpt.status as status, " +
		"cpt.id as id, " +
		"cpt.code as channel_pay_types_code, " +
		"c.currency_code as currency_code, " +
		"IFNULL(SUM(o.order_amount),0) as amount "

	tx := db.Table("ch_channel_pay_types cpt").
		Joins("join ch_pay_types pt on pt.code = cpt.pay_type_code").
		Joins("join ch_channels c on c.code = cpt.channel_code").
		Joins("left join tx_orders o on o.channel_pay_types_code = cpt.code "+
			"and o.trans_at > ? "+
			"and o.trans_at < ? "+
			"and o.status = 20 "+
			"and o.is_test = 0", req.StartAt, req.EndAt)

	if err = tx.Group("cpt.id").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.Select(selectX).
		Scopes(gormx.Paginate(*req)).Scopes(gormx.Sort(req.Orders)).
		Group("cpt.id").
		Find(&channelPayTypeQueryAmounts).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.ChannelPayTypeQueryAmountResponse{
		List:     channelPayTypeQueryAmounts,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
