package reportService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"gorm.io/gorm"
)

func InterMerchantReport(db *gorm.DB, req *types.MerchantReportQueryRequestX, ctx context.Context) (resp *types.MerchantReportQueryResponse, err error) {

	resp = &types.MerchantReportQueryResponse{}
	//var terms []string
	//var count int64
	var reportList []types.InternalMerchantReport
	endAt := utils.ParseTimeAddOneSecond(req.EndAt)

	db = db.Where("tx.`created_at` >= ?", req.StartAt)
	db = db.Where("tx.`created_at` < ?", endAt)
	db = db.Where("(ch.code is not null or chxf.code is not null )")
	db = db.Where("tx.is_test != '1'")

	if len(req.MerchantCode) > 0 {
		db = db.Where("tx.`merchant_code` like ?", "%"+req.MerchantCode+"%")
	}
	if len(req.ChannelCode) > 0 {
		db = db.Where("(ch.code like ? or chxf.code like ?' )", "%"+req.CurrencyCode+"%", "%"+req.ChannelCode+"%")
	}
	if len(req.ChannelName) > 0 {
		db = db.Where("(ch.`name` like ? or chxf.`name` like ?)", "%"+req.ChannelName+"%", "%"+req.ChannelName+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.TransactionType) > 0 {
		db = db.Where("tx.type = ?", req.TransactionType)
	} else {
		db = db.Where("tx.type in ('NC','ZF','DF','XF') ")
	}

	selectX := "tx.`merchant_code` AS merchant_code," +
		"CASE WHEN tx.type = 'XF' THEN chxf.`code` ELSE ch.`code` END AS channel_code, " + // 下發的渠道不同表
		"CASE WHEN tx.type = 'XF' THEN chxf.`name` ELSE ch.`name` END AS channel_name, " + // 下發的渠道不同表
		"tx.`type`            AS transaction_type," +
		"tx.currency_code     AS currency_code," +
		"pt.`name`            AS pay_type_name," +
		"mcr.fee              AS merchant_fee," +
		"mcr.handling_fee     AS merchant_handling_fee," +
		"cpt.fee              AS channel_fee," +
		"cpt.handling_fee     AS channel_handling_fee," +
		"SUM(tx.order_amount) AS order_amount," + //訂單總額
		"COUNT(*)             AS order_quantity," + //訂單數量
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.order_amount ELSE IF(tx.actual_amount != 0 ,tx.actual_amount,tx.order_amount) END ELSE 0 END) AS success_amount," + //成功總額
		"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS success_quantity," + //成功數量
		"floor(SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) / GREATEST(count(*),1)*100) AS success_rate," + //成功率
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.handling_fee ELSE ofp.transfer_handling_fee END ELSE 0 END) AS system_cost," + //系統成本 ps.下發單看(tx_order_channels.handling_fee) 其他單看(tx_orders_fee_profit.transfer_handling_fee)
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN ofp.profit_amount/cc.channel_count ELSE ofp.profit_amount END ELSE 0 END) AS system_profit " //系統利潤 ps.(下發要除該訂單的渠道數量)

	//selectTotal := "SUM(tx.order_amount) AS total_order_amount," + //訂單總額
	//	"COUNT(*) AS total_order_quantity," + //訂單數量
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.order_amount ELSE IF(tx.actual_amount != 0 ,tx.actual_amount,tx.order_amount) END ELSE 0 END) AS total_success_amount," + //成功總額
	//	"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS total_success_quantity," + //成功數量
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.handling_fee ELSE ofp.transfer_handling_fee END ELSE 0 END) AS total_cost," +
	//	"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN ofp.profit_amount/cc.channel_count ELSE ofp.profit_amount END ELSE 0 END) AS total_profit "

	//termWhere := strings.Join(terms, " AND ")

	tx := db.
		Table("tx_orders AS tx ").
		Joins("LEFT JOIN tx_orders_fee_profit ofp ON ofp.order_no = tx.order_no and ofp.merchant_code = '00000000' ").
		Joins("LEFT JOIN mc_merchant_channel_rate mcr ON mcr.merchant_code = tx.merchant_code and mcr.channel_pay_types_code = tx.channel_pay_types_code ").
		Joins("LEFT JOIN ch_channel_pay_types cpt ON cpt.`code` = tx.channel_pay_types_code ").
		Joins("LEFT JOIN ch_pay_types pt ON tx.pay_type_code = pt.`code` ").
		Joins("LEFT JOIN ch_channels ch ON tx.channel_code = ch.`code` ").
		Joins("LEFT JOIN tx_order_channels oc ON tx.order_no = oc.order_no "). // 下發用的渠道
		Joins("LEFT JOIN ch_channels chxf ON oc.channel_code = chxf.`code` "). // 下發用的渠道
		Joins("LEFT JOIN ( SELECT order_no, count(*) AS channel_count FROM tx_order_channels GROUP BY order_no ) cc ON cc.order_no = tx.order_no ")

	groupX := "tx.type, ch.code, chxf.code, pt.code, tx.merchant_code"

	if req.GroupType == "merchantCode" {
		groupX = "tx.merchant_code"
	} else if req.GroupType == "orderType" {
		groupX = "tx.type, tx.merchant_code"
	}

	//if err = tx.Select(selectTotal).Find(resp).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	//if err = tx.Group(groupX).Count(&count).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	if err = tx.Select(selectX).Group(groupX).Scopes(gormx.Sort(req.Orders)).
		Find(&reportList).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalOrderAmount := 0.0
	totalOrderQuantity := 0.0
	totalSuccessAmount := 0.0
	totalSuccessQuantity := 0.0
	totalCost := 0.0
	totalProfit := 0.0

	for _, report := range reportList {
		totalOrderAmount += report.OrderAmount
		totalOrderQuantity += report.OrderQuantity
		totalSuccessAmount += report.SuccessAmount
		totalSuccessQuantity += report.SuccessQuantity
		totalCost += report.SystemCost
		totalProfit += report.SystemProfit
	}

	//=============================================================================
	resp.List = reportList
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize
	resp.TotalOrderAmount = totalOrderAmount
	resp.TotalOrderQuantity = totalOrderQuantity
	resp.TotalSuccessAmount = totalSuccessAmount
	resp.TotalSuccessQuantity = totalSuccessQuantity
	resp.TotalCost = totalCost
	resp.TotalProfit = totalProfit
	return
}
