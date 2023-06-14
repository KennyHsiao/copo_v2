package orderrecordService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"gorm.io/gorm"
)

func IncomeExpenseQueryAll(db *gorm.DB, req types.IncomeExpenseQueryRequestX, ctx context.Context) (resp *types.IncomeExpenseQueryResponseX, err error) {
	var incomeExpenseRecords []types.IncomeExpenseRecordX
	var count int64
	var terms []string

	selectX := "r.*, " +
		"c.name as channel_name, " +
		"p.name_i18n->>'$." + req.Language + "' as pay_type_name"

	tx := db.Table("mc_merchant_balance_records as r").
		Joins("LEFT JOIN ch_channels c ON  r.channel_code = c.code").
		Joins("LEFT JOIN ch_pay_types p ON r.pay_type_code = p.code")

	if req.TransactionType == constants.TRANSACTION_TYPE_INTERNAL_CHARGE {
		// 查內充時 增加顯示訂單金額
		selectX += ",tx.order_amount as order_amount "
		tx.Joins("LEFT JOIN tx_orders tx ON  tx.order_no = r.order_no")
	}

	if len(req.MerchantCode) > 0 {
		tx = tx.Where("r.`merchant_code` = ?", req.MerchantCode)
		//terms = append(terms, fmt.Sprintf(" r.`merchant_code` = '%s'", req.MerchantCode))
	}
	if len(req.OrderNo) > 0 {
		tx = tx.Where("r.`order_no` = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf(" r.`order_no` = '%s'", req.OrderNo))
	}
	if len(req.OrderType) > 0 {
		tx = tx.Where("r.`order_type` = ?", req.OrderType)
		//terms = append(terms, fmt.Sprintf(" r.`order_type` = '%s'", req.OrderType))
	}
	if len(req.MerchantOrderNo) > 0 {
		tx = tx.Where("r.`merchant_order_no` = ?", req.MerchantOrderNo)
		//terms = append(terms, fmt.Sprintf(" r.`merchant_order_no` = '%s'", req.MerchantOrderNo))
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("r.`currency_code` = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf(" r.`currency_code` = '%s'", req.CurrencyCode))
	}
	if len(req.StartAt) > 0 {
		tx = tx.Where("r.`created_at` >= ?", req.StartAt)
		terms = append(terms, fmt.Sprintf(" r.`created_at` >= '%s'", req.StartAt))
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		tx = tx.Where("r.`created_at` < ?", endAt)
		//terms = append(terms, fmt.Sprintf(" r.`created_at` < '%s'", endAt))
	}
	if len(req.TransactionType) > 0 {
		tx = tx.Where("r.`transaction_type` = ?", req.TransactionType)
		//terms = append(terms, fmt.Sprintf(" r.`transaction_type` = '%s'", req.TransactionType))
	}
	if len(req.ChannelName) > 0 {
		tx = tx.Where("r.`name` like ?", "%"+req.ChannelName+"%")
		terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
	}
	if len(req.BalanceType) > 0 {
		tx = tx.Where("r.`balance_type` = ?", req.BalanceType)
		//terms = append(terms, fmt.Sprintf(" r.`balance_type`  = '%s'", req.BalanceType))
	}

	//term := strings.Join(terms, "AND")
	//tx.Where(term)

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.
		Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&incomeExpenseRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.IncomeExpenseQueryResponseX{
		List:     incomeExpenseRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
