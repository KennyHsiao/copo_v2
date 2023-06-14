package orderrecordService

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func ReceiptRecordQueryAll(db *gorm.DB, req types.ReceiptRecordQueryAllRequestX, isExcel bool, ctx context.Context) (resp *types.ReceiptRecordQueryAllResponseX, err error) {
	var receiptRecords []types.ReceiptRecordX
	var count int64
	//var terms []string

	selectX := "tx.id, " +
		"tx.merchant_code, " +
		"tx.order_no, " +
		"tx.merchant_order_no, " +
		"tx.channel_order_no, " +
		"tx.type, " +
		"tx.channel_pay_types_code, " +
		"tx.channel_code, " +
		"tx.pay_type_code, " +
		"tx.currency_code, " +
		"tx.balance_type, " +
		"tx.order_amount, " +
		"tx.actual_amount, " +
		"tx.merchant_bank_account, " +
		"tx.merchant_bank_name, " +
		"tx.merchant_account_name, " +
		"tx.channel_bank_account, " +
		"tx.channel_account_name, " +
		"tx.transfer_amount, " +
		"tx.transfer_handling_fee, " +
		"tx.handling_fee, " +
		"tx.fee, " +
		"tx.status, " +
		"tx.reason_type, " +
		"tx.call_back_status, " +
		"tx.is_merchant_callback, " +
		"tx.is_lock, " +
		"tx.is_test, " +
		"tx.memo, " +
		"tx.trans_at, " +
		"tx.created_at, " +
		"c.name as channel_name "

	tx := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code")

	if isExcel {
		selectX += ",p.name_i18n->>'$." + req.Language + "' as pay_type_name"
		tx.Joins("LEFT JOIN ch_pay_types p ON tx.pay_type_code = p.code")
	}

	if req.ReportType == "1" { // 代理報表

		tx.Joins("LEFT JOIN mc_merchants m ON m.code = tx.merchant_code ")

		selectX += ",m.agent_layer_code, " +
			"m.agent_parent_code "

		if req.JwtMerchantCode == "" {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}
		// 代理報表要顯示傭金收入
		tx.Joins("LEFT JOIN tx_orders_fee_profit fp ON tx.order_no = fp.order_no and fp.merchant_code = ?", req.JwtMerchantCode)
		selectX += ",fp.profit_amount as profit_amount "

		var merchants []types.Merchant
		var merchantCodes []string
		if merchants, err = model.NewMerchant(db).GetDescendantAgentsByCode(req.JwtMerchantCode, true); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		for _, m := range merchants {
			merchantCodes = append(merchantCodes, m.Code)
		}

		tx = tx.Where("tx.`merchant_code` in (?)", merchantCodes)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` in ('%s')", strings.Join(merchantCodes, "','")))

	} else if len(req.JwtMerchantCode) > 0 { // 一般報表
		tx = tx.Where(" tx.`merchant_code` = ?", req.JwtMerchantCode)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` = '%s'", req.JwtMerchantCode))
	}

	if len(req.MerchantCode) > 0 { // 一般報表
		tx = tx.Where(" tx.`merchant_code` = ?", req.MerchantCode)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` = '%s'", req.MerchantCode))
	}

	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.`order_no` = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf(" tx.`order_no` = '%s'", req.OrderNo))
	}
	if len(req.MerchantOrderNo) > 0 {
		tx = tx.Where("tx.`merchant_order_no` = ?", req.MerchantOrderNo)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_order_no` = '%s'", req.MerchantOrderNo))
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.`currency_code` = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf(" tx.`currency_code` = '%s'", req.CurrencyCode))
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.`trans_at` >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf(" tx.`trans_at` >= '%s'", req.StartAt))
		} else {
			tx = tx.Where("tx.`created_at` >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf(" tx.`created_at` >= '%s'", req.StartAt))
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.`trans_at` < ?", endAt)
			//terms = append(terms, fmt.Sprintf(" tx.`trans_at` < '%s'", endAt))
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.`created_at` < ?", endAt)
			//terms = append(terms, fmt.Sprintf(" tx.`created_at` < '%s'", endAt))
		}
	}
	if len(req.Status) > 0 {
		//tx = tx.Where("tx.`status` in (?)", strings.Join(req.Status, "','"))
		tx = tx.Where(fmt.Sprintf(" tx.`status` in ('%s') ", strings.Join(req.Status, "','")))
		//terms = append(terms, fmt.Sprintf(" tx.`status` in ('%s') ", strings.Join(req.Status, "','")))
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.`type` = ?", req.Type)
		//terms = append(terms, fmt.Sprintf(" tx.`type` = '%s'", req.Type))
	} else {
		tx = tx.Where("tx.`type` IN ('NC', 'ZF')")
		//terms = append(terms, fmt.Sprintf(" tx.`type` IN ('NC', 'ZF')"))
	}
	if len(req.PayTypeCode) > 0 {
		tx = tx.Where("tx.`pay_type_code` = ?", req.PayTypeCode)
		//terms = append(terms, fmt.Sprintf(" tx.`pay_type_code` = '%s'", req.PayTypeCode))
	}
	if len(req.ReasonTypes) > 1 {
		tx = tx.Where(fmt.Sprintf(" tx.`reason_type` in ('%s') ", strings.Join(req.ReasonTypes, "','")))
		//terms = append(terms, fmt.Sprintf(" tx.`reason_type` in ('%s') ", strings.Join(req.ReasonTypes, "','")))
	} else if len(req.ReasonTypes) == 1 {
		tx = tx.Where("tx.`reason_type` = ?", req.ReasonTypes[0])
		//terms = append(terms, fmt.Sprintf(" tx.`reason_type` = '%s'", req.ReasonTypes[0]))
	}
	if len(req.ChannelName) > 0 {
		tx = tx.Where("c.`name` like ?", "%"+req.ChannelName+"%")
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
	}

	//term := strings.Join(terms, "AND")
	//tx.Where(term)

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&receiptRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, _ := range receiptRecords {
		bankAccount := receiptRecords[i].MerchantBankAccount
		if len(bankAccount) > 5 {
			receiptRecords[i].MerchantBankAccountLastFive = bankAccount[len(bankAccount)-5:]
		} else {
			receiptRecords[i].MerchantBankAccountLastFive = bankAccount
		}
	}

	resp = &types.ReceiptRecordQueryAllResponseX{
		List:     receiptRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}

func ReceiptRecordTotalInfoBySuccess(db *gorm.DB, req types.ReceiptRecordQueryAllRequestX, ctx context.Context) (resp *types.ReceiptRecordTotalInfoResponse, err error) {

	selectStt := "sum(tx.actual_amount) as total_actual_amount," +
		"sum(tx.transfer_amount) as total_transfer_amount," +
		"sum(tx.transfer_handling_fee) as total_transfer_handling_fee"

	tx := db.Table("tx_orders AS tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code")

	// 这些总和只算成功单
	tx = tx.Where("tx.`status` = ?", constants.SUCCESS)
	if req.ReportType == "1" { // 代理報表
		if req.JwtMerchantCode == "" {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}
		// 代理報表要顯示傭金
		tx.Joins("LEFT JOIN tx_orders_fee_profit fp ON tx.order_no = fp.order_no and fp.merchant_code = ?", req.JwtMerchantCode)
		selectStt += ",sum(fp.profit_amount) as total_profit_amount "

		var merchants []types.Merchant
		var merchantCodes []string
		if merchants, err = model.NewMerchant(db).GetDescendantAgentsByCode(req.JwtMerchantCode, true); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		for _, m := range merchants {
			merchantCodes = append(merchantCodes, m.Code)
		}
		tx = tx.Where("tx.`merchant_code` in ?", merchantCodes)

	} else if len(req.JwtMerchantCode) > 0 { // 一般報表
		tx = tx.Where("tx.`merchant_code` = ?", req.JwtMerchantCode)
	}

	if len(req.MerchantCode) > 0 {
		tx = tx.Where("tx.`merchant_code` = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.`order_no` = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		tx = tx.Where("tx.`merchant_order_no` = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.`trans_at` >= ?", req.StartAt)
		} else {
			tx = tx.Where("tx.`created_at` >= ?", req.StartAt)
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.`trans_at` < ?", endAt)
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.`created_at` < ?", endAt)
		}
	}
	if len(req.Status) > 0 {
		tx = tx.Where(" tx.`status` in ? ", req.Status)
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.`type` = ?", req.Type)
	} else {
		tx = tx.Where("tx.`type` IN ('NC', 'ZF')")
	}
	if len(req.PayTypeCode) > 0 {
		tx = tx.Where("tx.`pay_type_code` = ?", req.PayTypeCode)
	}
	if len(req.ReasonTypes) > 1 {
		tx = tx.Where("tx.`reason_type` in ?", req.ReasonTypes)
	} else if len(req.ReasonTypes) == 1 {
		tx = tx.Where("tx.`reason_type` = ?", req.ReasonTypes[0])
	}
	if len(req.ChannelName) > 0 {
		tx = tx.Where(" c.`name` like ?", "%"+req.ChannelName+"%")
	}
	tx = tx.Where("tx.is_test != ?", constants.IS_TEST_YES)

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectStt).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func ReceiptRecordTotalOrderAmount(db *gorm.DB, req types.ReceiptRecordQueryAllRequestX, ctx context.Context) (totalOrderAmount float64, err error) {

	if req.ReportType == "1" { // 代理報表
		if req.JwtMerchantCode == "" {
			return 0, errorz.New(response.DATABASE_FAILURE)
		}
		var merchants []types.Merchant
		var merchantCodes []string
		if merchants, err = model.NewMerchant(db).GetDescendantAgentsByCode(req.JwtMerchantCode, true); err != nil {
			return 0, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		for _, m := range merchants {
			merchantCodes = append(merchantCodes, m.Code)
		}
		db = db.Where(" tx.`merchant_code` in ?", merchantCodes)

	} else if len(req.JwtMerchantCode) > 0 { // 一般報表
		db = db.Where(" tx.`merchant_code` = ?", req.JwtMerchantCode)
	}

	if len(req.MerchantCode) > 0 {
		db = db.Where("tx.`merchant_code` = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		db = db.Where("tx.`order_no` = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		db = db.Where("tx.`merchant_order_no` = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("tx.`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			db = db.Where("tx.`trans_at` >= ?", req.StartAt)
		} else {
			db = db.Where("tx.`created_at` >= ?", req.StartAt)
		}
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		if req.DateType == "2" {
			db = db.Where("tx.`trans_at` < ?", endAt)
		} else {
			db = db.Where("tx.`created_at` < ?", endAt)
		}
	}
	if len(req.Status) > 0 {
		db = db.Where(" tx.`status` in ? ", req.Status)
	}
	if len(req.Type) > 0 {
		db = db.Where("tx.`type` = ?", req.Type)
	} else {
		db = db.Where("tx.`type` IN ('NC', 'ZF')")
	}
	if len(req.PayTypeCode) > 0 {
		db = db.Where("tx.`pay_type_code` = ?", req.PayTypeCode)
	}
	if len(req.ReasonTypes) > 1 {
		db = db.Where("tx.`reason_type` in ? ", req.ReasonTypes)
	} else if len(req.ReasonTypes) == 1 {
		db = db.Where("tx.`reason_type` = ?", req.ReasonTypes[0])
	}
	if len(req.ChannelName) > 0 {
		db = db.Where("c.`name` like ?", "%"+req.ChannelName+"%")
	}
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)

	selectStt := "IFNULL(sum(tx.order_amount), 0) as total_order_amount"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectStt).Table("tx_orders AS tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
		Find(&totalOrderAmount).Error; err != nil {
		return 0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
