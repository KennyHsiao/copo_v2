package channelBalanceRecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBalanceRecordScheduleQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBalanceRecordScheduleQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBalanceRecordScheduleQueryAllLogic {
	return ChannelBalanceRecordScheduleQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBalanceRecordScheduleQueryAllLogic) ChannelBalanceRecordScheduleQueryAll(req *types.ChannelBalanceRecordQueryAllRequestX) (resp *types.ChannelBalanceRecordQueryAllResponseX, err error) {
	var list []types.ChannelBalanceRecordQuery
	var count int64
	db := l.svcCtx.MyDB

	selectX :=
		"c.currency_code           	as currency_code," +
			"c.name           			as channel_name," +
			"pt.name	        		as pay_type_name," +
			"cpt.id                		as id," +
			"cpt.code              		as code," +
			"cpt.channel_code      		as channel_code," +
			"cpt.pay_type_code     		as pay_type_code," +
			"cpt.fee      		   		as fee," +
			"cpt.handling_fee      		as handling_fee," +
			"cpt.max_internal_charge    as max_internal_charge," +
			"cpt.daily_tx_limit      	as daily_tx_limit," +
			"cpt.single_min_charge      as single_min_charge," +
			"cpt.single_max_charge      as single_max_charge," +
			"cpt.fixed_amount      		as fixed_amount," +
			"cpt.bill_date      		as bill_date," +
			"cpt.status      			as status," +
			"cpt.is_proxy				as is_proxy," +
			"cpt.device   				as device "

	if len(req.Code) > 0 {
		db = db.Where("cbr.code like ?", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		db = db.Where("cbr.name like ?", "%"+req.Name+"%")
	}
	if len(req.Time) > 0 {
		db = db.Where("cbr.time = ?", req.Time)
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("cbr.currency_code = ?", req.CurrencyCode)
	}
	if len(req.IsSuccess) > 0 {
		db = db.Where("cbr.is_success = ?", req.IsSuccess)
	}
	if len(req.StartAt) > 0 {
		db = db.Where("cbr.created_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		db = db.Where("cbr.created_at < ?", endAt)
	}

	if err = db.Table("ch_channel_pay_types cpt ").
		Joins("left join ch_channels c on cbr.code = c.code ").
		Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("ch_channel_balance_record cbr ").
		Select(selectX).
		Joins("left join ch_channels c on cbr.code = c.code ").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&list).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.ChannelBalanceRecordQueryAllResponseX{
		List:     list,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
