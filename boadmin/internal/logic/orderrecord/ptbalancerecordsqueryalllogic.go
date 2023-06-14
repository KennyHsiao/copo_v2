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

type PtBalanceRecordsQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPtBalanceRecordsQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) PtBalanceRecordsQueryAllLogic {
	return PtBalanceRecordsQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PtBalanceRecordsQueryAllLogic) PtBalanceRecordsQueryAll(req types.PtBalanceRecordsQueryRequestX) (resp *types.PtBalanceRecordsQueryResponseX, err error) {
	var ptBalanceRecordX []types.PtBalanceRecordX
	var count int64
	//var terms []string

	selectX := "r.*, " +
		"c.name as channel_name, " +
		"p.name_i18n->>'$." + req.Language + "' as pay_type_name," +
		"m.name as wallet_name"

	tx := l.svcCtx.MyDB.Table("mc_merchant_pt_balance_records as r").
		Joins("LEFT JOIN ch_channels c ON  r.channel_code = c.code").
		Joins("LEFT JOIN ch_pay_types p ON r.pay_type_code = p.code").
		Joins("LEFT JOIN mc_merchant_pt_balances m ON r.merchant_pt_balance_id = m.id")

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`merchant_code` = '%s'", req.MerchantCode))
		tx = tx.Where("r.merchant_code = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`order_no` = '%s'", req.OrderNo))
		tx = tx.Where("r.order_no = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`merchant_order_no` = '%s'", req.MerchantOrderNo))
		tx = tx.Where("r.merchant_order_no = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`currency_code` = '%s'", req.CurrencyCode))
		tx = tx.Where("r.currency_code = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`created_at` >= '%s'", req.StartAt))
		tx = tx.Where("r.created_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf(" r.`created_at` < '%s'", endAt))
		tx = tx.Where("r.created_at < ?", endAt)
	}
	if len(req.TransactionType) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`transaction_type` = '%s'", req.TransactionType))
		tx = tx.Where("r.transaction_type = ?", req.TransactionType)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		tx = tx.Where("c.name like ?", "%"+req.ChannelName+"%")
	}
	if len(req.PayTypeCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" r.`pay_type_code`  = '%s'", req.PayTypeCode))
		tx = tx.Where("r.pay_type_code = ?", req.PayTypeCode)
	}
	if len(req.WalletName) > 0 {
		tx = tx.Where("m.name = ?", req.WalletName)
	}
	//term := strings.Join(terms, "AND")
	//tx.Where(term)

	if err = tx.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.
		Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&ptBalanceRecordX).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.PtBalanceRecordsQueryResponseX{
		List:     ptBalanceRecordX,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
