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
	"gorm.io/gorm"
)

func DeductRecordQueryAll(db *gorm.DB, req types.DeductRecordQueryAllRequestX, isExcel bool, ctx context.Context) (resp *types.DeductRecordQueryAllResponseX, err error) {
	var deductRecords []types.DeductRecordX
	var count int64
	//var terms []string

	selectX := "DISTINCT tx.id, tx.merchant_code, tx.order_no, tx.merchant_order_no, tx.merchant_bank_name, " +
		"tx.merchant_bank_province, tx.merchant_account_name, tx.merchant_bank_account, tx.type, tx.order_amount, tx.fee, tx.handling_fee," +
		"tx.status, tx.memo, tx.error_note, tx.created_at, tx.trans_at, tx.reviewed_by, tx.source, tx.is_test, tx.channel_code," +
		"tx.transfer_handling_fee," +
		"m.agent_layer_code, " +
		"m.agent_parent_code " +
		", c.name as channel_name," +
		"tx.currency_code"

	tx := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
		Joins("LEFT JOIN mc_merchants m ON m.code = tx.merchant_code ").
		Joins("LEFT JOIN tx_order_channels oh ON oh.order_no = tx.order_no").
		Joins("LEFT JOIN ch_channels c2 ON oh.channel_code = c2.CODE ")

	//if isExcel == true {
	//	selectX += ", c.name as channel_name "
	//	tx.Joins("LEFT JOIN tx_order_channels oh ON oh.order_no = tx.order_no").
	//		Joins("LEFT JOIN ch_channels c2 ON oh.channel_code = c2.CODE ")
	//}

	if req.ReportType == "1" { // 代理報表
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
		tx = tx.Where("tx.`merchant_code` = ?", req.JwtMerchantCode)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` = '%s'", req.JwtMerchantCode))
	}

	if len(req.MerchantCode) > 0 { // 一般報表
		tx = tx.Where("tx.`merchant_code` = ?", req.MerchantCode)
		//terms = append(terms, fmt.Sprintf(" tx.`merchant_code` = '%s'", req.MerchantCode))
	}

	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.`order_no` = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf("tx.order_no = '%s'", req.OrderNo))
	}
	if len(req.MerchantOrderNo) > 0 {
		tx = tx.Where("tx.merchant_order_no = ?", req.MerchantOrderNo)
		//terms = append(terms, fmt.Sprintf("tx.merchant_order_no = '%s'", req.MerchantOrderNo))
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.currency_code = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.trans_at >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at >= '%s'", req.StartAt))
		} else {
			tx = tx.Where("tx.created_at >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at >= '%s'", req.StartAt))
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.trans_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at < '%s'", endAt))
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.created_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at < '%s'", endAt))
		}
	}
	if len(req.Status) > 0 {
		tx = tx.Where("tx.status = ?", req.Status)
		//terms = append(terms, fmt.Sprintf("tx.status = '%s'", req.Status))
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.type = ?", req.Type)
		//terms = append(terms, fmt.Sprintf("tx.Type = '%s'", req.Type))
	} else {
		tx = tx.Where("tx.type IN ('DF','XF')")
		//terms = append(terms, fmt.Sprintf("tx.type in ('DF','XF')"))
	}
	if len(req.Source) > 0 {
		tx = tx.Where("tx.Source = ?", req.Source)
		//terms = append(terms, fmt.Sprintf("tx.Source = '%s'", req.Source))
	}

	//term := strings.Join(terms, " AND ")

	if len(req.ChannelName) > 0 {

		tx = tx.Where("(c.name like ? OR c2.name like ?)", "%"+req.ChannelName+"%", "%"+req.ChannelName+"%").Group("tx.order_no")

		//terms = append(terms, fmt.Sprintf("(c.name like '%%%s%%' OR c2.name like '%%%s%%')", req.ChannelName, req.ChannelName))
		//term = strings.Join(terms, " AND ")
		//term = term + fmt.Sprintf(" GROUP BY tx.order_no")
	}

	//tx.Where(term)

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Distinct("tx.order_no").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&deductRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, record := range deductRecords {
		deductRecords[i].CreatedAt = utils.ParseTime(record.CreatedAt)
	}

	resp = &types.DeductRecordQueryAllResponseX{
		List:     deductRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}

func DeductRecordTotalInfo(db *gorm.DB, req types.DeductRecordQueryAllRequestX, ctx context.Context) (resp *types.DeductRecordTotalInfoResponse, err error) {

	selectX := "tx.*"
	tx := db.Table("tx_orders tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
		Joins("LEFT JOIN tx_order_channels oh ON oh.order_no = tx.order_no").
		Joins("LEFT JOIN ch_channels c2 ON oh.channel_code = c2.CODE ")

	if req.ReportType == "1" { // 代理報表
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
		tx = tx.Where(" tx.`merchant_code` in ?", merchantCodes)

	} else if len(req.JwtMerchantCode) > 0 { // 一般報表
		tx = tx.Where(" tx.`merchant_code` = ?", req.JwtMerchantCode)
	}

	if len(req.MerchantCode) > 0 {
		tx = tx.Where("tx.merchant_code = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.order_no = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		tx = tx.Where("tx.merchant_order_no = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.trans_at >= ?", req.StartAt)
		} else {
			tx = tx.Where("tx.created_at >= ?", req.StartAt)
		}
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		if req.DateType == "2" {
			tx = tx.Where("tx.trans_at < ?", endAt)
		} else {
			tx = tx.Where("tx.created_at < ?", endAt)
		}
	}
	if len(req.Status) > 0 {
		tx = tx.Where("tx.status = ?", req.Status)
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.Type = ?", req.Type)
	} else {
		tx = tx.Where("tx.Type in ('DF','XF')")
	}
	if len(req.Source) > 0 {
		tx = tx.Where("tx.Source = ?", req.Source)
	}
	if len(req.ChannelName) > 0 {
		tx = tx.Where("(c.name like ? OR c2.name like ?)", "%"+req.ChannelName+"%", "%"+req.ChannelName+"%")
	}
	tx = tx.Where("tx.is_test != ?", constants.IS_TEST_YES)

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	var order []types.OrderD
	if err = tx.Distinct(selectX).Find(&order).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	var totalOrderAmount float64
	var totalTransferHandlingFee float64
	var totalProfitAmount float64
	for _, os := range order {
		totalProfitAmount = utils.FloatAdd(totalProfitAmount, os.ProfitAmount)
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, os.OrderAmount)
		totalTransferHandlingFee = utils.FloatAdd(totalTransferHandlingFee, os.TransferHandlingFee)
	}

	resp = &types.DeductRecordTotalInfoResponse{
		TotalOrderAmount:         totalOrderAmount,
		TotalTransferHandlingFee: totalTransferHandlingFee,
		TotalProfitAmount:        totalProfitAmount,
	}
	return
}
