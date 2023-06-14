package reportService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

func GetQueryTodayTime(nowTime time.Time) (startTimeStr, endTimeStr string) {
	startTime := nowTime.AddDate(0, 0, -1).Format("2006-01-02")
	startTimeStr = fmt.Sprint(startTime) + " 16:00:00"
	endTime := nowTime.Format("2006-01-02")
	endTimeStr = fmt.Sprintf(endTime) + " 16:00:00"

	return startTimeStr, endTimeStr
}

/**依输入的i做日期加减，渠得今日以前6天日期（含今天）*/
func GetWeeklyDate(i int, nowTime time.Time) (startTime, endTime, showTime string) {
	var startTimeStr string
	var endTimeStr string
	showTime = nowTime.AddDate(0, 0, -i).Format("2006-01-02")
	st := nowTime.AddDate(0, 0, -i-1).Format("2006-01-02")
	startTimeStr = fmt.Sprint(st) + " 16:00:00"
	et := nowTime.AddDate(0, 0, -i).Format("2006-01-02")
	endTimeStr = fmt.Sprintf(et) + " 16:00:00"

	return startTimeStr, endTimeStr, showTime
}

/** 取得当前全部商户馀额*/
func GetTotalMerchantBalance(db *gorm.DB, currencyCode string, ctx context.Context) (resp *types.TotalMerchantBalances, err error) {
	selectX := "SUM(CASE WHEN balance_type = 'DFB' THEN balance END) AS df_balances," +
		"SUM(CASE WHEN balance_type = 'XFB' THEN balance END) AS xf_balances"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Table("mc_merchant_balances").Select(selectX).
		Where("currency_code = ?", currencyCode).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}

/**
取得某一时间段的某一币别的某一订单类型的总订单金额，
orderType没传就是查询全部类型
*/
func GetOneOrderTypeTotalAmount(db *gorm.DB, orderType, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {
	resp = float64(0.0)
	//var terms []string
	//terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	//terms = append(terms, fmt.Sprintf("status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("is_test != '%s'", constants.IS_TEST_YES))
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("status = ?", constants.SUCCESS)
	db = db.Where("is_test != ?", constants.IS_TEST_YES)
	if len(orderType) > 0 {
		//terms = append(terms, fmt.Sprintf("type = '%s'", orderType))
		db = db.Where("type = ?", orderType)
		if orderType == "XF" {
			//terms = append(terms, fmt.Sprintf("created_at >= '%s'", startTime))
			//terms = append(terms, fmt.Sprintf("created_at < '%s'", endTime))
			db = db.Where("created_at >= ?", startTime)
			db = db.Where("created_at < ?", endTime)
		} else {
			//terms = append(terms, fmt.Sprintf("trans_at >= '%s'", startTime))
			//terms = append(terms, fmt.Sprintf("trans_at < '%s'", endTime))
			db = db.Where("trans_at >= ?", startTime)
			db = db.Where("trans_at < ?", endTime)
		}
	}

	//term := strings.Join(terms, " AND ")

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Table("tx_orders").Select("COALESCE(SUM(order_amount),0)").Find(&resp).Error; err != nil {
		return 0.0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return resp, nil
}

/**系统利润*/
func GetTotalSystemProfitAmount(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {
	resp = float64(0.0)
	//var terms []string
	//terms = append(terms, fmt.Sprintf("tp.created_at >= '%s'", startTime))
	//terms = append(terms, fmt.Sprintf("tp.created_at < '%s'", endTime))
	//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", currencyCode))
	//terms = append(terms, fmt.Sprintf("tp.merchant_code = '00000000'"))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	db = db.Where("tp.created_at >= ?", startTime)
	db = db.Where("tp.created_at < ?", endTime)
	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	//term := strings.Join(terms, " AND ")

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Table("tx_orders_fee_profit tp").
		Select("COALESCE(SUM(tp.profit_amount),0)").
		Joins("join tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return 0.0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return resp, nil
}

/**取得收益細節*/
func GetPerDayIncomeDetail(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp *types.IncomeDetailResponse, err error) {
	//var terms []string
	//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", currencyCode))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	////terms = append(terms, fmt.Sprintf("tp.created_at >= '%s'", startTime))
	////terms = append(terms, fmt.Sprintf("tp.created_at < '%s'", endTime))
	//terms = append(terms, fmt.Sprintf("tp.merchant_code = '00000000'"))

	//term := strings.Join(terms, " AND ")
	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")

	selectX := "COALESCE(SUM(CASE WHEN tx.type = 'NC' AND tx.trans_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS nc_rate," +
		"COALESCE(SUM(CASE WHEN tx.type = 'ZF' AND tx.trans_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS zf_rate," +
		"COALESCE(SUM(CASE WHEN tx.type = 'DF' AND tx.trans_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS df_rate," +
		"COALESCE(SUM(CASE WHEN tx.type = 'XF' AND tx.created_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS xf_handling_fee"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func GetPerDayIncomeDetail2(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp []types.TotalPayout, err error) {
	//var terms []string
	//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", currencyCode))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	////terms = append(terms, fmt.Sprintf("tp.created_at >= '%s'", startTime))
	////terms = append(terms, fmt.Sprintf("tp.created_at < '%s'", endTime))
	//terms = append(terms, fmt.Sprintf("tp.merchant_code = '00000000'"))
	//terms = append(terms, fmt.Sprintf("tp.created_at BETWEEN '%s' AND '%s'", startTime, endTime))
	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tp.created_at BETWEEN ? AND ?", startTime, endTime)

	//term := strings.Join(terms, " AND ")

	selectX :=
		"date(DATE_ADD(tx.trans_at, INTERVAL 8 HOUR)) as date," +
			"COALESCE(SUM(CASE WHEN tx.type = 'DF' AND tx.trans_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS proxy_pay_handling_fee," +
			"COALESCE(SUM(CASE WHEN tx.type = 'XF' AND tx.created_at BETWEEN '" + startTime + "' AND '" + endTime + "' THEN tp.profit_amount END),0) AS withdraw_handling_fee"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Group("date").Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func GetPerDayIncomeDetailNc(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {

	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tx.type = 'NC'")
	db = db.Where("tx.trans_at >= ?", startTime)
	db = db.Where("tx.trans_at < ?", endTime)

	selectX := "COALESCE(SUM(tp.profit_amount),0) AS nc_rate"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return 0.000, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func GetPerDayIncomeDetailZf(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {

	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tx.type = 'ZF'")
	db = db.Where("tx.trans_at >= ?", startTime)
	db = db.Where("tx.trans_at < ?", endTime)

	selectX := "COALESCE(SUM(tp.profit_amount),0) AS zf_rate"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return 0.000, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func GetPerDayIncomeDetailDf(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {

	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tx.type = 'DF'")
	db = db.Where("tx.created_at >= ?", startTime)
	db = db.Where("tx.created_at < ?", endTime)

	selectX := "COALESCE(SUM(tp.profit_amount),0) AS df_rate"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return 0.000, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func GetPerDayIncomeDetailXf(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp float64, err error) {

	db = db.Where("tx.currency_code = ?", currencyCode)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tp.merchant_code = '00000000'")
	db = db.Where("tx.type = 'XF'")
	db = db.Where("tx.trans_at >= ?", startTime)
	db = db.Where("tx.trans_at < ?", endTime)

	selectX := "COALESCE(SUM(tp.profit_amount),0) AS xf_handling_fee"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Select(selectX).Table("tx_orders_fee_profit tp").
		Joins("JOIN tx_orders tx on tp.order_no = tx.order_no").
		Find(&resp).Error; err != nil {
		return 0.000, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

//func GetWeeklyOrderAmount(db *gorm.DB, currencyCode, startTime, endTime string) (resp []types.WeeklyTransDetail, err error) {
//
//	var terms []string
//	terms = append(terms, fmt.Sprintf("1=1"))
//	terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
//	terms = append(terms, fmt.Sprintf("status = '%s'", constants.SUCCESS))
//	terms = append(terms, fmt.Sprintf("is_test != '%s'", constants.IS_TEST_YES))
//	terms = append(terms, fmt.Sprintf("trans_at >= '%s'", startTime))
//	terms = append(terms, fmt.Sprintf("trans_at < '%s'", endTime))
//
//	term := strings.Join(terms, " AND ")
//
//	selectX := "date(DATE_ADD(trans_at, INTERVAL 8 HOUR)) as date," +
//		"SUM(if(tx.type='ZF',order_amount,0)) as ZfTotalAmount," +
//		"SUM(if(tx.type='NC',order_amount,0)) as NcTotalAmount," +
//		"SUM(if(tx.type='DF',order_amount,0)) as DfTotalAmount," +
//		"SUM(if(tx.type='XF' and tx.created_at >='"+startTime+"' and  tx.created_at < '"+endTime+"',order_amount,0)) as XfTotalAmount"
//	groupX := "date(DATE_ADD(trans_at, INTERVAL 8 HOUR))"
//
//	if err = db.Table("tx_orders tx").Select(selectX).Where(term).Group(groupX).Find(&resp).Error; err != nil {
//		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
//	}
//	return resp, nil
//}

func GetWeeklyOrderAmount(db *gorm.DB, currencyCode, startTime, endTime string, ctx context.Context) (resp []types.WeeklyTransDetail, err error) {
	//var terms []string
	//terms = append(terms, fmt.Sprintf("1=1"))
	//terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	//terms = append(terms, fmt.Sprintf("status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("is_test != '%s'", constants.IS_TEST_YES))
	//terms = append(terms, fmt.Sprintf("trans_at >= '%s'", startTime))
	//terms = append(terms, fmt.Sprintf("trans_at < '%s'", endTime))
	//
	//term := strings.Join(terms, " AND ")
	db = db.Where("1=1")
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("status = ?", constants.SUCCESS)
	db = db.Where("is_test != ?", constants.IS_TEST_YES)
	db = db.Where("trans_at >= ?", startTime)
	db = db.Where("trans_at < ?", endTime)

	selectX :=
		"SUM(if(tx.type='ZF',order_amount,0)) as ZfTotalAmount," +
			"SUM(if(tx.type='NC',order_amount,0)) as NcTotalAmount," +
			"SUM(if(tx.type='DF',order_amount,0)) as DfTotalAmount," +
			"SUM(if(tx.type='XF' and tx.created_at >='" + startTime + "' and  tx.created_at < '" + endTime + "',order_amount,0)) as XfTotalAmount"
	//groupX := "date(DATE_ADD(trans_at, INTERVAL 8 HOUR))"

	if ctx != nil {
		db = db.WithContext(ctx)
	}

	if err = db.Table("tx_orders tx").Select(selectX).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return resp, nil
}
