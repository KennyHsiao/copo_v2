package orderrecord

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

type FrozenRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFrozenRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) FrozenRecordQueryAllLogic {
	return FrozenRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FrozenRecordQueryAllLogic) FrozenRecordQueryAll(req types.FrozenRecordQueryAllRequestX) (resp *types.FrozenRecordQueryAllResponseX, err error) {
	var frozenRecords []types.FrozenRecordX
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` = '%s'", req.MerchantCode))
		db = db.Where("tx.`merchant_code` = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`order_no` = '%s'", req.OrderNo))
		db = db.Where("tx.`order_no` = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_order_no` = '%s'", req.MerchantOrderNo))
		db = db.Where("tx.`merchant_order_no` = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`currency_code` = '%s'", req.CurrencyCode))
		db = db.Where("tx.`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`updated_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`updated_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf(" tx.`updated_at` < '%s'", endAt))
		db = db.Where("tx.`updated_at` < ?", endAt)
	}
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`type` = '%s'", req.Type))
		db = db.Where("tx.`type` = ?", req.Type)
	}
	if len(req.PayTypeCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" tx.`pay_type_code` = '%s'", req.PayTypeCode))
		db = db.Where("tx.`pay_type_code` = ?", req.PayTypeCode)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		db = db.Where("c.`name` like ?", "%"+req.ChannelName+"%")
	}

	//terms = append(terms, fmt.Sprintf(" tx.`status` = '%s'", "31"))
	db = db.Where("tx.status = ?", "31")
	//term := strings.Join(terms, "AND")

	selectX := "tx.*, " +
		"c.name as channel_name, " +
		"p.name_i18n->>'$." + req.Language + "' as pay_type_name"

	db2 := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
		Joins("LEFT JOIN ch_pay_types p ON tx.pay_type_code = p.code")
	db2.Count(&count)

	if err = db2.Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&frozenRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.FrozenRecordQueryAllResponseX{
		List:     frozenRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
