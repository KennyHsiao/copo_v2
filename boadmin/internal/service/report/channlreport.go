package reportService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strings"
)

type InterTotalInfo struct {
	SystemCost   float64 `json:"system_cost"`
	SystemProfit float64 `json:"system_profit"`
}

func InterChannelReport(db *gorm.DB, req *types.ChannelReportQueryRequestX, ctx context.Context) (resp *types.ChannelReportQueryresponse, err error) {

	resp = &types.ChannelReportQueryresponse{}

	var reportList []types.InternalChannelReport

	endAt := utils.ParseTimeAddOneSecond(req.EndAt)

	channelNameMap, err := getChannelNameMap(db, req.CurrencyCode)
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	payTypeNameMap, err := getPayTypeNameMap(db, req.CurrencyCode)
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	db = db.Where("tx.`created_at` >= ?", req.StartAt)
	db = db.Where("tx.`created_at` < ?", endAt)
	db = db.Where("((tx.channel_code is not null and tx.channel_code != '') or (oc.channel_code is not null and oc.channel_code != ''))")
	db = db.Where("tx.is_test != '1'")

	if len(req.ChannelCode) > 0 {
		db = db.Where("(tx.channel_code like ? or oc.channel_code like ? )", "%"+req.ChannelCode+"%", "%"+req.ChannelCode+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.TransactionType) > 0 {
		db = db.Where("tx.type = ?", req.TransactionType)
	} else {
		db = db.Where("tx.type in ('NC','ZF','DF','XF') ")
	}
	if len(req.PayType) > 0 {
		db = db.Where("tx.pay_type_code = ?", req.PayType)
	}

	selectX :=
		"date(DATE_ADD(tx.created_at, INTERVAL 8 HOUR)) as date," +
			"CASE WHEN tx.type = 'XF' THEN oc.channel_code ELSE tx.channel_code END AS channel_code," + // 下發的渠道不同表
			"tx.`type`            AS transaction_type," +
			"tx.currency_code     AS currency_code," +
			"tx.pay_type_code     AS pay_type_code," +
			"CASE WHEN tx.change_type = '1' AND tx.type = 'XF' THEN cptxf.fee ELSE cpt.fee END AS fee," +
			"CASE WHEN tx.change_type = '1' AND tx.type = 'XF' THEN cptxf.handling_fee ELSE cpt.handling_fee END AS handling_fee," +
			"SUM( CASE WHEN tx.type = 'XF' AND tx.change_type = '1' THEN oc.order_amount ELSE tx.order_amount END) AS order_amount," + //訂單總額
			"COUNT(*)             AS order_quantity, " + //訂單數量
			"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' AND oc.`status` = '20' THEN oc.order_amount  ELSE IF ( tx.actual_amount != 0, tx.actual_amount, tx.order_amount ) END ELSE 0 END ) AS success_amount," + //成功總額
			"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS success_quantity," + //成功數量
			"floor(SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) / GREATEST(count(*),1)*100) AS success_rate," + //成功率
			"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN CASE WHEN tx.change_type = '1' AND oc.`status` = '20' THEN  IF(cptxf.is_rate = '1',oc.order_amount * oc.fee /100 + oc.handling_fee,oc.handling_fee) ELSE oc.handling_fee END ELSE ofp.transfer_handling_fee END ELSE 0 END ) AS system_cost," + //系統成本
			"SUM( CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN CASE WHEN tx.change_type = '1' AND oc.`status` = '20' THEN IF(cptxf.is_rate = '1', oc.transfer_handling_fee - oc.order_amount * oc.fee /100 + oc.handling_fee,oc.handling_fee) ELSE ofp.profit_amount / cc.channel_count END ELSE ofp.profit_amount END ELSE 0  END ) AS system_profit" //系統利潤

	//termWhere := strings.Join(terms, " AND ")

	tx := db.
		Table("tx_orders AS tx ").
		Joins("LEFT JOIN tx_orders_fee_profit ofp ON ofp.order_no = tx.order_no and ofp.merchant_code = '00000000' ").
		Joins("LEFT JOIN ch_channel_pay_types cpt ON cpt.`code` = tx.channel_pay_types_code ").
		Joins("LEFT JOIN tx_order_channels oc ON tx.order_no = oc.order_no "). // 下發用的渠道
		Joins("LEFT JOIN ch_channel_pay_types cptxf ON cptxf.`code` = CONCAT(oc.channel_code,'DF') ").
		Joins("LEFT JOIN ( SELECT order_no, count(*) AS channel_count FROM tx_order_channels GROUP BY order_no ) cc ON cc.order_no = tx.order_no ")

	groupX := "tx.type, tx.channel_code, oc.channel_code, tx.pay_type_code"
	if req.ReportType == "date" {
		groupX += ", date(DATE_ADD(tx.created_at, INTERVAL 8 HOUR))"
	} else {
		tx.Order("FIELD(tx.`type`,'NC','ZF','DF','XF')")
	}

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	//if err = tx.Group(groupX).Count(&count).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	if err = tx.Select(selectX).Group(groupX).Scopes(gormx.Sort(req.Orders)).
		Find(&reportList).Error; err != nil {
		if ctx != nil {
			logx.WithContext(ctx).Error(err.Error())
		}
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, report := range reportList {
		reportList[i].Date = strings.Split(report.Date, "T")[0]
		reportList[i].ChannelName = channelNameMap[report.ChannelCode]
		reportList[i].PayTypeName = payTypeNameMap[report.PayTypeCode]
	}

	//=============================================================================
	resp.List = reportList

	return
}

func InterChannelReportTotal(db *gorm.DB, req *types.ChannelReportQueryRequestX, ctx context.Context) (resp *types.ChannelReportTotalResponse, err error) {

	resp = &types.ChannelReportTotalResponse{}
	//var terms []string

	endAt := utils.ParseTimeAddOneSecond(req.EndAt)
	//terms = append(terms, fmt.Sprintf("tx.`created_at` >= '%s'", req.StartAt))
	//terms = append(terms, fmt.Sprintf("tx.`created_at` < '%s'", endAt))
	//terms = append(terms, " (ch.code is not null or chxf.code is not null ) ") // 這條件是撇除沒有渠道的下發單
	//terms = append(terms, " tx.is_test != '1' ")                               // 撇除測試單
	db = db.Where("tx.`created_at` >= ?", req.StartAt)
	db = db.Where("tx.`created_at` < ?", endAt)
	db = db.Where("((tx.channel_code is not null and tx.channel_code != '') or (oc.channel_code is not null and oc.channel_code != ''))")
	db = db.Where("tx.is_test != '1'")

	if len(req.ChannelCode) > 0 {
		db = db.Where("(tx.channel_code like ? or oc.channel_code like ? )", "%"+req.ChannelCode+"%", "%"+req.ChannelCode+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.TransactionType) > 0 {
		db = db.Where("tx.type = ?", req.TransactionType)
	} else {
		db = db.Where("tx.type in ('NC','ZF','DF','XF') ")
	}
	if len(req.PayType) > 0 {
		db = db.Where("tx.pay_type_code = ?", req.PayType)
	}

	selectTotal := "SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN oc.handling_fee ELSE ofp.transfer_handling_fee END ELSE 0 END) AS total_cost," +
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' THEN ofp.profit_amount/cc.channel_count ELSE ofp.profit_amount END ELSE 0 END) AS total_profit," +
		"SUM( CASE WHEN tx.type = 'XF' AND tx.change_type = '1' THEN oc.order_amount ELSE tx.order_amount END) AS total_order_amount," + //訂單總額
		"COUNT(*) AS total_order_quantity, " + //訂單數量
		"SUM(CASE WHEN tx.`status` = '20' THEN CASE WHEN tx.type = 'XF' AND oc.`status` = '20' THEN oc.order_amount  ELSE IF ( tx.actual_amount != 0, tx.actual_amount, tx.order_amount ) END ELSE 0 END ) AS total_success_amount," + //成功總額
		"SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) AS total_success_quantity," + //成功數量
		"floor(SUM(CASE WHEN tx.`status` = '20' THEN 1 ELSE 0 END) / GREATEST(count(*),1)*100) AS total_success_rate " //成功率

	tx := db.
		Table("tx_orders AS tx ").
		Joins("LEFT JOIN tx_orders_fee_profit ofp ON ofp.order_no = tx.order_no and ofp.merchant_code = '00000000' ").
		Joins("LEFT JOIN tx_order_channels oc ON tx.order_no = oc.order_no "). // 下發用的渠道
		Joins("LEFT JOIN ( SELECT order_no, count(*) AS channel_count FROM tx_order_channels GROUP BY order_no ) cc ON cc.order_no = tx.order_no ")

	groupX := "tx.type, tx.channel_code, oc.channel_code, tx.pay_type_code"
	if req.ReportType == "date" {
		groupX += ", date(DATE_ADD(tx.created_at, INTERVAL 8 HOUR))"
	} else {
		tx.Order("FIELD(tx.`type`,'NC','ZF','DF','XF')")
	}

	if err = tx.Select(selectTotal).Find(resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func getChannelNameMap(db *gorm.DB, currencyCode string) (channelNameMap map[string]string, err error) {
	var channels []types.ChannelData
	channelNameMap = map[string]string{}
	err = db.Table("ch_channels").Where("currency_code = ?", currencyCode).Find(&channels).Error
	for _, channel := range channels {
		channelNameMap[channel.Code] = channel.Name
	}
	return
}

func getPayTypeNameMap(db *gorm.DB, currencyCode string) (payTypeNameMap map[string]string, err error) {
	var payTypes []types.PayType
	payTypeNameMap = map[string]string{}
	err = db.Table("ch_pay_types").Where("currency like ?", "%"+currencyCode+"%").Find(&payTypes).Error
	for _, payType := range payTypes {
		payTypeNameMap[payType.Code] = payType.Name
	}
	return
}
