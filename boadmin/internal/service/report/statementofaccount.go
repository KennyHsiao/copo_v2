package reportService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"gorm.io/gorm"
)

type InternalChangeInfo struct {
	TotalNum                 int
	TotalOrderAmount         float64
	TotalTransferHandlingFee float64
}

type ProxyPayInfo struct {
	TotalNum           int
	TotalOrderAmount   float64
	ChannelHandlingFee float64
	SystemCommission   float64
	AgentCommission    float64
}

func InterChargeCheckBill(db *gorm.DB, req *types.PayCheckBillQueryRequestX, ctx context.Context) (resp *types.InternalChargeCheckBillQueryResponse, err error) {
	var internalChargeCheckBills []types.InternalChargeCheckBill
	//var terms []string

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("tx.merchant_code = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		db = db.Where("c.`name` like ?", "%"+req.ChannelName+"%")
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`trans_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` < '%s'", endAt))
		db = db.Where("tx.`trans_at` < ?", endAt)
	}
	//terms = append(terms, fmt.Sprintf("tx.type = '%s'", constants.ORDER_TYPE_NC))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	////terms = append(terms, fmt.Sprintf("tp.merchant_code != '00000000'"))
	db = db.Where("tx.type = ?", constants.ORDER_TYPE_NC)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	//term := strings.Join(terms, " AND ")

	/**** 此版本已tx_orders_fee_profit為主表，代財務確認不需用此版本後刪除 ****/
	//selectX := "tp.merchant_code," +
	//	"c.`name` as channel_name," +
	//	"SUM(CASE WHEN tp.merchant_code = tx.merchant_code then 1 else 0 end) as total_num," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tx.actual_amount) else 0 end as total_order_amount," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tx.transfer_handling_fee) else 0 end as total_transfer_handling_fee," +
	//	"SUM(tp.profit_amount) as agent_commission," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tp2.profit_amount) else 0 end as system_commission"
	//
	//tx := db.Table("tx_orders_fee_profit tp").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp3 on tp3.order_no = tp.order_no and tp3.merchant_code = tp.agent_parent_code").
	//	Joins("LEFT JOIN tx_orders as tx on tx.order_no = tp.order_no").
	//	Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
	//	Where(term).Group("tp.merchant_code, tx.channel_code")
	/**** end ****/

	selectX := "tx.merchant_code," +
		"c.`name` as channel_name," +
		"count(*) as total_num," +
		"SUM(tx.order_amount) as total_order_amount," +
		"SUM(tx.transfer_handling_fee) as total_transfer_handling_fee," +
		"SUM(tp.transfer_handling_fee - tp2.transfer_handling_fee - tp2.profit_amount) as agent_commission," +
		"SUM(tp2.profit_amount) as system_commission"

	tx := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code").
		Joins("LEFT JOIN tx_orders_fee_profit tp on tx.order_no = tp.order_no and tx.merchant_code = tp.merchant_code").
		Joins("LEFT JOIN tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
		Group("tx.merchant_code, tx.channel_code")

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectX).Scopes(gormx.Sort(req.Orders)).Find(&internalChargeCheckBills).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalNum := 0
	var totalOrderAmount float64
	var totalTransferHandlingFee float64
	var systemCommission float64
	var AgentCommission float64
	for _, bill := range internalChargeCheckBills {
		totalNum += bill.TotalNum
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, bill.TotalOrderAmount)
		totalTransferHandlingFee = utils.FloatAdd(totalTransferHandlingFee, bill.TotalTransferHandlingFee)
		systemCommission = utils.FloatAdd(systemCommission, bill.SystemCommission)
		AgentCommission = utils.FloatAdd(AgentCommission, bill.AgentCommission)
	}

	//interChangeInfo := InternalChangeInfo{}
	//if err = db.Table("tx_orders as tx").
	//	Select("count(*) AS total_num, SUM(order_amount) AS total_order_amount, SUM(transfer_handling_fee) AS total_transfer_handling_fee").
	//	Where(term).
	//	Find(&interChangeInfo).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	resp = &types.InternalChargeCheckBillQueryResponse{
		List:                     internalChargeCheckBills,
		TotalNum:                 totalNum,
		TotalOrderAmount:         totalOrderAmount,
		TotalTransferHandlingFee: totalTransferHandlingFee,
		SystemCommission:         systemCommission,
		AgentCommission:          AgentCommission,
	}
	return
}

func WithdrawCheckBill(db *gorm.DB, req *types.WithdrawCheckBillQueryRequestX, ctx context.Context) (resp *types.WithdrawCheckBillQueryResponse, err error) {
	var withdrawCheckBills []types.WithdrawCheckBill
	//var terms []string

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("tx.merchant_code = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.`created_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`created_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.`created_at` < '%s'", endAt))
		db = db.Where("tx.`created_at` < ?", endAt)
	}
	//terms = append(terms, fmt.Sprintf("tx.type = '%s'", constants.ORDER_TYPE_XF))
	db = db.Where("tx.type = ?", constants.ORDER_TYPE_XF)
	//term := strings.Join(terms, " AND ")

	selectX := "tx.merchant_code," +
		"count(*) as withdraw_total_num," +
		"SUM(case when tx.status = '20' then 1 else 0 end) AS total_success_num," +
		"SUM(case when tx.status = '1' then 1 else 0 end) AS total_process_num," +
		"COALESCE(SUM( tx.order_amount),0) as total_order_amount," +
		"COALESCE(SUM(case when tx.status = '20' then tx.handling_fee end),0) as total_handling_fee"
		//"case when temp.total_order_amount IS NULL then 0 else temp.total_order_amount end AS total_order_amount," +
		//"SUM(case when temp.total_handling_fee IS NULL then 0 else temp.total_handling_fee end) AS total_handling_fee"

	tx := db.Table("tx_orders as tx").Group("tx.merchant_code")

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectX).Scopes(gormx.Sort(req.Orders)).Find(&withdrawCheckBills).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalNum := 0
	totalSuccessNum := 0
	totalProcessNum := 0
	var totalOrderAmount float64
	var totalHandlingFee float64
	for _, bill := range withdrawCheckBills {
		totalNum = totalNum + bill.WithdrawTotalNum
		totalSuccessNum = totalSuccessNum + bill.TotalSuccessNum
		totalProcessNum = totalProcessNum + bill.TotalProcessNum
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, bill.TotalOrderAmount)
		totalHandlingFee = utils.FloatAdd(totalHandlingFee, bill.TotalHandlingFee)
	}

	resp = &types.WithdrawCheckBillQueryResponse{
		List:             withdrawCheckBills,
		WithdrawTotalNum: totalNum,
		TotalSuccessNum:  totalSuccessNum,
		TotalProcessNum:  totalProcessNum,
		TotalOrderAmount: totalOrderAmount,
		TotalHandlingFee: totalHandlingFee,
	}

	return
}

func ProxyPayCheckBill(db *gorm.DB, req *types.PayCheckBillQueryRequestX, ctx context.Context) (resp *types.ProxyPayCheckBillQueryResponse, err error) {
	var proxyPayCheckBills []types.ProxyPayCheckBill
	//var terms []string

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("tx.merchant_code = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		db = db.Where(" c.`name` like ?", "%"+req.ChannelName+"%")
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`trans_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` < '%s'", endAt))
		db = db.Where("tx.`trans_at` < ?", endAt)
	}
	//terms = append(terms, fmt.Sprintf("tx.type = '%s'", constants.ORDER_TYPE_DF))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	////terms = append(terms, fmt.Sprintf("tp.merchant_code != '00000000'"))
	db = db.Where("tx.type = ?", constants.ORDER_TYPE_DF)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	//term := strings.Join(terms, " AND ")

	/***** 此版本已tx_orders_fee_profit為主表，代財務確認不需用此版本後刪除 *****/
	//selectX := "tp.merchant_code," +
	//	"c.name as channel_name," +
	//	"sum(case when tp.merchant_code = tx.merchant_code then 1 else 0 end) as total_num," +
	//	"case when tp.merchant_code = tx.merchant_code then sum(tx.order_amount) else 0 end as total_order_amount," +
	//	"case when tp.merchant_code = tx.merchant_code then sum(tx.handling_fee) else 0 end as total_handling_fee," +
	//	"case when tp.merchant_code = tx.merchant_code then sum(cpt.handling_fee) else 0 end as channel_handling_fee," +
	//	"sum(tp.profit_amount) as agent_commission," +
	//	"case when tp.merchant_code = tx.merchant_code then sum(tp2.profit_amount) else 0 end as system_commission"
	//
	//tx := db.Table("tx_orders_fee_profit tp").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp3 on tp3.order_no = tp.order_no and tp3.merchant_code = tp.agent_parent_code").
	//	Joins("LEFT JOIN tx_orders as tx on tx.order_no = tp.order_no").
	//	Joins("LEFT JOIN ch_channels c on tx.channel_code = c.code").
	//	Joins("LEFT JOIN ch_channel_pay_types cpt on tx.pay_type_code = cpt.pay_type_code AND tx.channel_code = cpt.channel_code").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
	//	Where(term).Group("tp.merchant_code, tx.channel_code")
	/***** end *****/

	selectX := "tx.merchant_code," +
		"c.name as channel_name," +
		"count(*) as total_num," +
		"sum(tx.order_amount) as total_order_amount," +
		"sum(tx.transfer_handling_fee) as total_handling_fee," +
		"sum(cpt.handling_fee) as channel_handling_fee," +
		"sum(tp.transfer_handling_fee - tp2.transfer_handling_fee - tp2.profit_amount) as agent_commission," +
		"sum(tp2.profit_amount) as system_commission"

	tx := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c on tx.channel_code = c.code").
		Joins("LEFT JOIN ch_channel_pay_types cpt on tx.pay_type_code = cpt.pay_type_code AND tx.channel_code = cpt.channel_code").
		Joins("LEFT JOIN tx_orders_fee_profit tp on tx.order_no = tp.order_no and tx.merchant_code = tp.merchant_code").
		Joins("LEFT JOIN tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
		Group("tx.merchant_code, tx.channel_code")

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectX).Scopes(gormx.Sort(req.Orders)).Find(&proxyPayCheckBills).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//selectX2 := "count(*) as total_num," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tx.order_amount) else 0 end as total_order_amount," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(cpt.handling_fee) else 0 end as channel_handling_fee," +
	//	"SUM(tp.profit_amount) as agent_commission," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tp2.profit_amount) else 0 end as system_commission"
	//
	//tx2 := db.Table("tx_orders_fee_profit tp").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp3 on tp3.order_no = tp.order_no and tp3.merchant_code = tp.agent_parent_code").
	//	Joins("LEFT JOIN tx_orders as tx on tx.order_no = tp.order_no").
	//	Joins("LEFT JOIN ch_channel_pay_types cpt on tx.pay_type_code = cpt.pay_type_code and tx.channel_code = cpt.channel_code").
	//	Joins("LEFT JOIN tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
	//	Where(term)

	//proxyPayInfo := ProxyPayInfo{}
	//if err = tx2.Select(selectX2).Find(&proxyPayInfo).Error; err != nil {
	//	return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	var totalNum int
	var totalOrderAmount float64
	var channelHandlingFee float64
	var systemCommission float64
	var agentCommission float64
	var totalHandlingFee float64
	var proxyPayHandlingFee float64
	for i, bill := range proxyPayCheckBills {
		totalNum += bill.TotalNum
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, bill.TotalOrderAmount)
		channelHandlingFee = utils.FloatAdd(channelHandlingFee, bill.ChannelHandlingFee)
		systemCommission = utils.FloatAdd(systemCommission, bill.SystemCommission)
		agentCommission = utils.FloatAdd(agentCommission, bill.AgentCommission)
		totalHandlingFee =
			utils.FloatAdd(totalHandlingFee, utils.FloatAdd(bill.ChannelHandlingFee, utils.FloatAdd(bill.SystemCommission, bill.AgentCommission)))
		proxyPayCheckBills[i].TotalHandlingFee = utils.FloatAdd(proxyPayHandlingFee, bill.TotalHandlingFee)
	}

	resp = &types.ProxyPayCheckBillQueryResponse{
		List:               proxyPayCheckBills,
		TotalNum:           totalNum,
		TotalOrderAmount:   totalOrderAmount,
		TotalHandlingFee:   totalHandlingFee,
		ChannelHandlingFee: channelHandlingFee,
		SystemCommission:   systemCommission,
		AgentCommission:    agentCommission,
	}
	return
}

func PayCheckBill(db *gorm.DB, req *types.PayCheckBillQueryRequestX, ctx context.Context) (resp *types.PayCheckBillQueryResponse, err error) {
	var payCheckBills []types.PayCheckBill
	//var terms []string

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("tx.merchant_code = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		db = db.Where(" c.`name` like ?", "%"+req.ChannelName+"%")
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`trans_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` < '%s'", endAt))
		db = db.Where("tx.`trans_at` < ?", endAt)
	}
	//terms = append(terms, fmt.Sprintf("tx.type = '%s'", constants.ORDER_TYPE_ZF))
	//terms = append(terms, fmt.Sprintf("tx.status in ('%s','%s')", constants.SUCCESS, constants.FROZEN))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	//terms = append(terms, fmt.Sprintf("tx.reason_type != '%s'", constants.ORDER_REASON_TYPE_RECOVER))
	////terms = append(terms, fmt.Sprintf("tp.merchant_code != '00000000'"))
	db = db.Where("tx.type = ?", constants.ORDER_TYPE_ZF)
	db = db.Where("tx.status in (?,?)", constants.SUCCESS, constants.FROZEN)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("tx.reason_type != ?", constants.ORDER_REASON_TYPE_RECOVER)
	//term := strings.Join(terms, " AND ")

	/***** 此版本已tx_orders_fee_profit為主表，代財務確認不需用此版本後刪除 *****/
	//selectX := "tp.merchant_code," +
	//	"c.name as channel_name," +
	//	"pt.name as pay_type_name," +
	//	"SUM(CASE WHEN tp.merchant_code = tx.merchant_code then 1 else 0 end) as total_num," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tx.actual_amount) else 0 end as total_order_amount," +
	//	//"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tx.transfer_handling_fee + tp2.profit_amount + (tp.transfer_handling_fee - tp2.transfer_handling_fee - tp2.profit_amount)) else 0 end as total_handling_fee," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tp2.transfer_handling_fee) else 0 end as pay_total_transfer_handling_fee," +
	//	"CASE WHEN tp.merchant_code = tx.merchant_code then SUM(tp2.profit_amount) else 0 end as system_commission," +
	//	"SUM(tp.profit_amount) as agent_commission," +
	//	"case when SUM(mcr.transfer_amount) is NULL then 0 else SUM(mcr.transfer_amount) END as adjustment_amount"
	//
	//tx := db.Table("tx_orders_fee_profit tp").
	//	Joins("left join tx_orders_fee_profit tp3 on tp3.order_no = tp.order_no and tp3.merchant_code = tp.agent_parent_code").
	//	Joins("left join tx_orders as tx on tx.order_no = tp.order_no").
	//	Joins("left join ch_channels c on tx.channel_code = c.`code`").
	//	Joins("left join ch_pay_types pt on tx.pay_type_code = pt.code").
	//	Joins("left join tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
	//	Joins("left join mc_merchant_balance_records mcr on tx.merchant_code = mcr.merchant_code and tx.order_no = mcr.order_no and mcr.transaction_type = '20'").
	//	Where(term).Group("tp.merchant_code, tx.channel_code, tx.pay_type_code")
	/***** end *****/

	selectX := "tx.merchant_code," +
		"c.name as channel_name," +
		"pt.name as pay_type_name," +
		"count(*) as total_num," +
		"SUM(tx.actual_amount) as total_order_amount," +
		"SUM(tx.transfer_handling_fee + tp2.profit_amount + (tp.transfer_handling_fee - tp2.transfer_handling_fee - tp2.profit_amount)) as total_handling_fee," +
		"SUM(tx.transfer_handling_fee) as pay_total_transfer_handling_fee," +
		"SUM(tp2.profit_amount) as system_commission," +
		"SUM(tp.transfer_handling_fee - tp2.transfer_handling_fee - tp2.profit_amount) as agent_commission"
		//"case when SUM(mcr.transfer_amount) is NULL then 0 else SUM(mcr.transfer_amount) END as adjustment_amount"

	tx := db.Table("tx_orders as tx").
		Joins("left join ch_channels c on tx.channel_code = c.`code`").
		Joins("left join ch_pay_types pt on tx.pay_type_code = pt.code").
		Joins("left join tx_orders_fee_profit tp on tx.order_no = tp.order_no and tx.merchant_code = tp.merchant_code").
		Joins("left join tx_orders_fee_profit tp2 on tx.order_no = tp2.order_no and tp2.merchant_code = '00000000'").
		Group("tx.merchant_code, tx.channel_code, tx.pay_type_code")

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectX).Scopes(gormx.Sort(req.Orders)).Find(&payCheckBills).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	totalNum := 0
	var totalOrderAmount float64
	var systemCommission float64
	var AgentCommission float64
	var Adjustment_amount float64
	var TotalHandlingFee float64
	var PayTotalHandlingFee float64
	var totalSystemCost float64

	for i, bill := range payCheckBills {
		totalNum = totalNum + bill.TotalNum
		payCheckBills[i].TotalHandlingFee = utils.FloatAdd(bill.PayTotalTransferHandlingFee, utils.FloatAdd(bill.SystemCommission, bill.AgentCommission))
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, bill.TotalOrderAmount)
		systemCommission = utils.FloatAdd(systemCommission, bill.SystemCommission)
		AgentCommission = utils.FloatAdd(AgentCommission, bill.AgentCommission)
		Adjustment_amount = utils.FloatAdd(Adjustment_amount, bill.AdjustmentAmount)
		TotalHandlingFee = utils.FloatAdd(TotalHandlingFee, utils.FloatAdd(bill.PayTotalTransferHandlingFee, utils.FloatAdd(bill.SystemCommission, bill.AgentCommission)))
		PayTotalHandlingFee = utils.FloatAdd(PayTotalHandlingFee, bill.PayTotalTransferHandlingFee)
	}

	totalSystemCost = utils.FloatSub(PayTotalHandlingFee, utils.FloatAdd(systemCommission, AgentCommission))

	resp = &types.PayCheckBillQueryResponse{
		List:                        payCheckBills,
		TotalNum:                    totalNum,
		TotalOrderAmount:            totalOrderAmount,
		SystemCommission:            systemCommission,
		AgentCommission:             AgentCommission,
		TotalHandlingFee:            TotalHandlingFee,
		PayTotalTransferHandlingFee: PayTotalHandlingFee,
		AdjustmentAmount:            Adjustment_amount,
		TotalSystemCost:             totalSystemCost,
	}

	return
}

func AppropriationCheckBill(db *gorm.DB, req *types.AppropriationCheckBillQueryRequest, ctx context.Context) (resp *types.AppropriationCheckBillQueryResponse, err error) {
	var Appropriats []types.AppropriationCheckBillQuery
	//var terms []string
	db2 := db
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf(" c.`name` like '%%%s%%'", req.ChannelName))
		db = db.Where(" c.`name` like ?", "%"+req.ChannelName+"%")
		db2 = db2.Where(" c.`name` like ?", "%"+req.ChannelName+"%")
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
		db2 = db2.Where("tx.currency_code = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` >= '%s'", req.StartAt))
		db = db.Where("tx.`trans_at` >= ?", req.StartAt)
		db2 = db2.Where("tx.`trans_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.`trans_at` < '%s'", endAt))
		db = db.Where("tx.`trans_at` < ?", endAt)
		db2 = db2.Where("tx.`trans_at` < ?", endAt)
	}

	//terms = append(terms, fmt.Sprintf("tx.type = '%s'", constants.ORDER_TYPE_BK))
	//terms = append(terms, fmt.Sprintf("tx.status = '%s'", constants.SUCCESS))
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))
	db = db.Where("tx.type = ?", constants.ORDER_TYPE_BK)
	db = db.Where("tx.status = ?", constants.SUCCESS)
	db = db.Where("tx.is_test != ?", constants.IS_TEST_YES)
	db2 = db2.Where("tx.type = ?", constants.ORDER_TYPE_BK)
	db2 = db2.Where("tx.status = ?", constants.SUCCESS)
	db2 = db2.Where("tx.is_test != ?", constants.IS_TEST_YES)
	//term := strings.Join(terms, " AND ")

	selectX := "c.name AS channel_name," +
		"SUM(tx.order_amount) AS appropriation_amount," +
		"count(*) AS appropriation_count," +
		"SUM(tx.transfer_handling_fee) AS appropriation_handling_fee"

	tx := db.Table("tx_orders as tx").
		Joins("left join ch_channels c on tx.channel_code = c.code").
		Group("c.name")

	if ctx != nil {
		tx = tx.WithContext(ctx)
	}

	if err = tx.Select(selectX).Find(&Appropriats).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	data := struct {
		TotalAmount      float64 `json:"total_amount"`
		TotalCount       int     `json:"total_count"`
		TotalHandlingFee float64 `json:"total_handling_fee"`
	}{}

	selectY := "SUM(tx.order_amount) AS total_amount," +
		"count(*) AS total_count," +
		"SUM(tx.transfer_handling_fee) AS total_handling_fee"

	tx2 := db2.Table("tx_orders as tx").
		Joins("left join ch_channels c on tx.channel_code = c.code")

	if ctx != nil {
		tx2 = tx2.WithContext(ctx)
	}

	if err = tx2.Select(selectY).Find(&data).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.AppropriationCheckBillQueryResponse{
		List:             Appropriats,
		TotalHandlingFee: data.TotalHandlingFee,
		TotalAmount:      data.TotalAmount,
		TotalCount:       data.TotalCount,
	}

	return
}
