package ordersService

import (
	"com.copo/bo_service/boadmin/internal/config"
	"com.copo/bo_service/boadmin/internal/model"
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/bo"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/gozzle"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

type ProxyPayOrderCallBackRequest struct {
	ProxyPayOrderNo     string  `json:"proxyPayOrderNo"`     // 平台订单号
	ChannelOrderNo      string  `json:"channelOrderNo"`      //渠道商回复单号
	ChannelResultAt     string  `json:"channelResultAt"`     //渠道商回复日期  //(YYYYMMDDhhmmss)
	ChannelResultStatus string  `json:"channelResultStatus"` //渠道商回复处理状态  //(Dior渠道商范例：状态 0处理中，1成功，2失败) */
	ChannelResultNote   string  `json:"channelResultNote"`   //渠道商回复处理备注
	Amount              float64 `json:"amount"`              //代付金额
	ChannelCharge       float64 `json:"channelCharge"`       //渠道商成本
	UpdatedBy           string  `json:"updatedBy"`           //更新人员
}

type Body struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Trace   string      `json:"trace"`
}

func WithdrawOrderCreate(db *gorm.DB, req []types.OrderWithdrawCreateRequestX, ctx context.Context, svcCtx *svc.ServiceContext) (resp *types.OrderWithdrawCreateResponse, err error) {
	var orders = req
	var size = len(orders)
	var orderNos []string
	orderNoMap := make(map[string]string)

	for size != len(orderNoMap) {
		//产生单号
		orderNo := model.GenerateOrderNo("XF")
		if _, isExist := orderNoMap[orderNo]; !isExist {
			orderNoMap[orderNo] = orderNo
		}
	}

	for _, v := range orderNoMap {
		orderNos = append(orderNos, v)
	}

	var handlingFee float64
	systemRate := types.SystemRate{}
	merchantCurrency := &types.MerchantCurrency{}
	merchantCode := orders[0].MerchantCode
	userAccount := orders[0].UserAccount
	var currency = orders[0].CurrencyCode
	var terms []string

	terms = append(terms, fmt.Sprintf("currency_code = '%s'", currency))
	term := strings.Join(terms, " AND ")

	// 取得商户下发手续费
	if err = db.Table("mc_merchant_currencies").Where(term).Where(" merchant_code = ?", merchantCode).Find(&merchantCurrency).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 取得系统下发上下限资料
	if err = db.Table("bs_system_rate").Where(term).Take(&systemRate).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorz.New(response.SYSTEM_RATE_NOT_SET, err.Error())
	} else if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	// 判断有无设定下发上下限与手续费
	if systemRate.MinWithdrawCharge <= 0 {
		return nil, errorz.New(response.MER_WITHDRAW_MIN_LIMIT_NOT_SET)
	} else if systemRate.MaxWithdrawCharge <= 0 {
		return nil, errorz.New(response.MER_WITHDRAW_MAX_LIMIT_NOT_SET)
	} else if merchantCurrency.WithdrawHandlingFee <= 0 {
		if systemRate.WithdrawHandlingFee <= 0 {
			return nil, errorz.New(response.MER_WITHDRAW_CHARGE_NOT_SET)
		} else {
			handlingFee = systemRate.WithdrawHandlingFee
		}
	} else if merchantCurrency.WithdrawHandlingFee > 0 {
		handlingFee = merchantCurrency.WithdrawHandlingFee
	}

	var idxs []string
	var errs []string
	for i, order := range orders {
		ux := model.NewBankBlockAccount(db)
		isBlockReceive, err3 := ux.CheckIsBlockAccount(order.MerchantBankAccount)
		if isBlockReceive {
			logx.WithContext(ctx).Errorf("错误:此账号为黑名单，账号: ", order.MerchantBankAccount)
			idxs = append(idxs, order.Index)
			continue
		} else if err3 != nil {
			logx.WithContext(ctx).Errorf("錯誤:判断黑名单错误，Index : %v, err : %v", order.Index, err3.Error())
			errs = append(errs, order.Index)
			continue
		}

		//验证银行卡号(必填)(必须为数字)(长度必须在10~22码)
		isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, order.MerchantBankAccount)
		currencyCode := order.CurrencyCode
		if currencyCode == constants.CURRENCY_THB {
			if order.MerchantBankAccount == "" || len(order.MerchantBankAccount) < 10 || len(order.MerchantBankAccount) > 16 || !isMatch2 {
				logx.WithContext(ctx).Error("銀行卡號檢查錯誤，需10-16碼內：", order.MerchantBankAccount)
				return nil, errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+order.MerchantBankAccount)
			}
		} else if currencyCode == constants.CURRENCY_CNY {
			if order.MerchantBankAccount == "" || len(order.MerchantBankAccount) < 13 || len(order.MerchantBankAccount) > 22 || !isMatch2 {
				logx.WithContext(ctx).Error("銀行卡號檢查錯誤，需13-22碼內：", order.MerchantBankAccount)
				return nil, errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+order.MerchantBankAccount)
			}
		}

		transAmount := utils.FloatAdd(order.OrderAmount, handlingFee)
		if systemRate.MaxWithdrawCharge < transAmount {
			//下发金额超过上限
			logx.WithContext(ctx).Errorf("錯誤:下發金額超過上限，Index : , err : %d", order.Index)
			errs = append(errs, order.Index)
			continue
		} else if systemRate.MinWithdrawCharge > transAmount {
			//下发金额未达下限
			logx.WithContext(ctx).Errorf("錯誤:下發金額未達下限，Index : %d, err : %d", order.Index)
			errs = append(errs, order.Index)
			continue
		}

		var errRpc error
		var res *transaction.WithdrawOrderResponse
		orderNo := orderNos[i]
		withdrawOrderReq := &transaction.WithdrawOrderRequest{
			MerchantCode:         merchantCode,
			UserAccount:          userAccount,
			MerchantAccountName:  order.MerchantAccountName,
			MerchantBankeAccount: order.MerchantBankAccount,
			MerchantBankNo:       order.MerchantBankNo,
			MerchantBankName:     order.MerchantBankName,
			MerchantBankProvince: order.MerchantBankProvince,
			MerchantBankCity:     order.MerchantBankCity,
			CurrencyCode:         order.CurrencyCode,
			OrderAmount:          order.OrderAmount,
			OrderNo:              orderNo,
			HandlingFee:          handlingFee,
			Source:               constants.UI,
			PtBalanceId:          order.PtBalanceId,
		}
		res, errRpc = svcCtx.TransactionRpc.WithdrawOrderTransaction(ctx, withdrawOrderReq)

		if err := transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
			MerchantCode: merchantCode,
			//MerchantOrderNo: req.OrderNo,
			OrderNo:       orderNos[i],
			LogType:       constants.MERCHANT_REQUEST,
			LogSource:     constants.PLATEFORM_XF,
			TxOrderSource: constants.UI,
			TxOrderType:   constants.ORDER_TYPE_XF,
			Content:       withdrawOrderReq,
			TraceId:       trace.SpanContextFromContext(ctx).TraceID().String(),
		}); err != nil {
			logx.WithContext(ctx).Errorf("写入交易日志错误:%s", err)
		}

		if errRpc != nil {
			logx.WithContext(ctx).Error("UI WithdrawOrder Tranaction rpcResp error:%s", errRpc.Error())
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else if res.Code != response.API_SUCCESS {
			logx.WithContext(ctx).Errorf("UI WithdrawOrder Tranaction error Code:%s, Message:%s", res.Code, res.Message)
			return nil, errorz.New(res.Code, res.Message)
		} else if res.Code == response.API_SUCCESS {
			logx.WithContext(ctx).Infof("頁面下发提单rpc完成，单号: %v", res.OrderNo)
		}
	}

	//for i, order := range orders {
	//	// 初始化订单
	//	var newOrder types.OrderX
	//	newOrder.OrderNo = orderNos[i]
	//	newOrder.Type = constants.ORDER_TYPE_XF
	//	newOrder.Status = constants.WAIT_PROCESS
	//	newOrder.MerchantCode = merchantCode
	//	newOrder.CreatedBy = userAccount
	//	newOrder.UpdatedBy = userAccount
	//	newOrder.BalanceType = "XFB"
	//	newOrder.CurrencyCode = currency
	//	newOrder.TransferAmount = utils.FloatAdd(order.OrderAmount, merchantCurrency.WithdrawHandlingFee)
	//	newOrder.OrderAmount = order.OrderAmount
	//	newOrder.MerchantBankAccount = order.MerchantBankAccount
	//	newOrder.MerchantBankNo = order.MerchantBankNo
	//	newOrder.MerchantBankProvince = order.MerchantBankProvince
	//	newOrder.MerchantBankCity = order.MerchantBankCity
	//	newOrder.MerchantAccountName = order.MerchantAccountName
	//	newOrder.MerchantOrderNo = order.MerchantOrderNo
	//	newOrder.IsLock = "0" //是否锁定状态 (0=否;1=是) 预设否
	//	newOrder.PageUrl = order.PageUrl
	//	newOrder.HandlingFee = merchantCurrency.WithdrawHandlingFee
	//
	//	if len(order.NotifyUrl) > 0 {
	//		newOrder.NotifyUrl = order.NotifyUrl
	//	}
	//	if len(order.MerchantOrderNo) > 0 {
	//		newOrder.MerchantOrderNo = order.MerchantOrderNo
	//	}
	//	if len(order.MerchantBankName) > 0 {
	//		newOrder.MerchantBankName = order.MerchantBankName
	//	}
	//
	//	tx := db.Begin()
	//	//判斷黑名單，收款與付款都要判斷
	//	//API提单须改为失败单
	//	ux := model.NewBankBlockAccount(db)
	//	isBlockReceive, err3 := ux.CheckIsBlockAccount(order.MerchantBankAccount)
	//	if isBlockReceive &&  orderSource == constants.API{
	//		logx.Infof("交易账户%s-%s在黑名单内，使用0元假扣款", newOrder.MerchantAccountName, newOrder.MerchantBankNo)
	//		newOrder.ErrorType = constants.ERROR6_BANK_ACCOUNT_IS_BLACK //交易账户为黑名单
	//		newOrder.ErrorNote = constants.BANK_ACCOUNT_IS_BLACK        //失败原因：黑名单交易失败
	//		newOrder.Status = constants.PROXY_PAY_FAIL                  //状态:失败
	//		newOrder.Fee = 0                                            //写入本次手续费(未发送到渠道的交易，都设为0元)
	//		newOrder.HandlingFee = 0
	//		newOrder.TransAt = types.JsonTime{}.New()
	//		logx.Infof("商户 %s，下发订单 %#v ，交易账户为黑名单", newOrder.MerchantCode, newOrder)
	//	} else if isBlockReceive {
	//		logx.Errorf("错误:此账号为黑名单，账号: ", order.MerchantBankAccount)
	//		idxs = append(idxs, order.Index)
	//		tx.Rollback()
	//		continue
	//	} else if err3 != nil {
	//		logx.Errorf("錯誤:判断黑名单错误，Index : %v, err : %v", order.Index, err3.Error())
	//		errs = append(errs, order.Index)
	//		tx.Rollback()
	//		continue
	//	}
	//
	//	if merchantCurrency.MaxWithdrawCharge < newOrder.TransferAmount {
	//		//下发金额超过上限
	//		if orderSource == constants.UI {
	//			logx.Errorf("錯誤:下發金額超過上限，Index : , err : %d", order.Index)
	//			errs = append(errs, order.Index)
	//			tx.Rollback()
	//			continue
	//		} else {
	//			tx.Rollback()
	//			return nil, errorz.New(response.WITHDRAW_AMT_EXCEED_MAX_LIMIT)
	//		}
	//	} else if merchantCurrency.MinWithdrawCharge > newOrder.TransferAmount {
	//		//下发金额未达下限
	//		if orderSource == constants.UI {
	//			logx.Errorf("錯誤:下發金額未達下限，Index : %d, err : %d", order.Index)
	//			errs = append(errs, order.Index)
	//			tx.Rollback()
	//			continue
	//		} else {
	//			tx.Rollback()
	//			return nil, errorz.New(response.WITHDRAW_AMT_NOT_REACH_MIN_LIMIT)
	//		}
	//	}
	//
	//	// 新增收支记录，更新商户余额
	//	updateBalance := types.UpdateBalance{
	//		MerchantCode:    merchantCode,
	//		CurrencyCode:    currency,
	//		OrderNo:         newOrder.OrderNo,
	//		OrderType:       newOrder.Type,
	//		TransactionType: "11",
	//		BalanceType:     newOrder.BalanceType,
	//		TransferAmount:  -newOrder.TransferAmount,
	//		CreatedBy:       userAccount,
	//	}
	//	var merchantBalanceRecord types.MerchantBalanceRecord
	//	merchantBalanceRecord, err = merchantbalanceservice.UpdateBalance(tx, updateBalance)
	//	if err != nil {
	//		if orderSource == constants.UI {
	//			logx.Errorf("錯誤:新增收支紀錄錯誤，Index : %d, err : %d", order.Index, err.Error())
	//			errs = append(errs, order.Index)
	//			tx.Rollback()
	//			continue
	//		} else {
	//			tx.Rollback()
	//			return nil, err
	//		}
	//	}
	//
	//	newOrder.CallBackStatus = "1"  // 回调状态
	//	newOrder.Source = orderSource // 訂單來源: 1:平台 2:API
	//	newOrder.TransferHandlingFee = utils.FloatSub(math.Abs(merchantBalanceRecord.TransferAmount), order.OrderAmount)
	//	newOrder.BeforeBalance = merchantBalanceRecord.BeforeBalance
	//	newOrder.Balance = merchantBalanceRecord.AfterBalance
	//
	//	if err = tx.Table("tx_orders").Create(&newOrder).Error; err != nil {
	//		if orderSource == constants.UI {
	//			logx.Errorf("錯誤:新增下發訂單錯誤，Index : %d, err : %d", order.Index, err.Error())
	//			errs = append(errs, order.Index)
	//			tx.Rollback()
	//			continue
	//		} else {
	//			tx.Rollback()
	//			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//		}
	//	}
	//	err = tx.Commit().Error
	//	if err != nil {
	//	}
	//	newOrders = append(newOrders, newOrder)
	//}

	// 新增下發訂單資料
	//if err = db.Table("tx_orders").CreateInBatches(newOrders, len(newOrders)).Error; err != nil {
	//	db.Rollback()
	//	return nil,errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	//for _, order := range newOrders {
	//	//记录订单历程
	//	orderAction := types.OrderAction{
	//		OrderNo:     order.OrderNo,
	//		Action:      "PLACE_ORDER",
	//		UserAccount: userAccount,
	//		Comment:     "",
	//	}
	//	if err = model.NewOrderAction(db).CreateOrderAction(&types.OrderActionX{
	//		OrderAction: orderAction,
	//	}); err != nil {
	//		tx.Rollback()
	//		return nil,err
	//	}
	//}

	resp = &types.OrderWithdrawCreateResponse{
		Index: idxs,
		Errs:  errs,
	}

	return
}

func WithdrawApiCallBack(db *gorm.DB, req types.OrderX, ctx context.Context) error {
	var orderX types.OrderX
	var merchant types.Merchant
	// 確認單號是否存在
	if err := db.Table("tx_orders").Where("merchant_order_no = ?", req.MerchantOrderNo).Take(&orderX).Error; err != nil {
		logx.WithContext(ctx).Errorf("UI下发回调错误: 查无订单。商户订单号: %v", req.MerchantOrderNo)
		return errorz.New(response.INVALID_ORDER_NO, err.Error())
	}

	// 取得商戶密鑰
	if err := db.Table("mc_merchants").Where("code = ?", req.MerchantCode).Take(&merchant).Error; err != nil {
		logx.WithContext(ctx).Errorf("下UI发回调错误: 查无商戶。商户号: %v", req.MerchantCode)
		return errorz.New(response.INVALID_MERCHANT_ID, err.Error())
	}

	// 状态 0：处理中，1：成功，2：失败
	var orderStatus = "0"
	// 訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	if orderX.Status == "20" {
		orderStatus = "1"
	} else if orderX.Status == "30" {
		orderStatus = "2"
	}

	resp := vo.WithdrawCallBackVO{
		MerchantId:  orderX.MerchantCode,
		OrderNo:     orderX.MerchantOrderNo,
		OrderAmount: fmt.Sprintf("%.2f", orderX.OrderAmount),
		OrderTime:   orderX.CreatedAt.Format("20060102150405"),
		ReviewTime:  orderX.TransAt.Time().Format("20060102150405"),
		Fee:         fmt.Sprintf("%.2f", orderX.TransferHandlingFee),
		OrderStatus: orderStatus,
		DiorOrderNo: orderX.OrderNo,
	}
	logx.WithContext(ctx).Infof("UI下发回调商户参数，resp : %+v", resp)
	resp.Sign = utils.SortAndSign2(resp, merchant.ScrectKey)

	//新增交易日志
	if err := transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
		MerchantCode:    orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		ChannelOrderNo:  orderX.ChannelOrderNo,
		LogType:         constants.CALLBACK_TO_MERCHANT,
		LogSource:       constants.API_XF,
		TxOrderSource:   orderX.Source,
		TxOrderType:     orderX.Type,
		Content:         resp,
		TraceId:         trace.SpanContextFromContext(ctx).TraceID().String(),
	}); err != nil {
		logx.WithContext(ctx).Errorf("写入交易日志错误:%s", err)
	}

	// 通知商戶
	span := trace.SpanFromContext(ctx)
	res, err := gozzle.Post(orderX.NotifyUrl).Timeout(10).Trace(span).JSON(resp)
	if err != nil || res.Status() != 200 {
		contentStrut := struct {
			StatusCode int
			Body       string
			ErrorMsg   string
		}{}
		if err != nil {
			contentStrut.ErrorMsg = err.Error()
		} else {
			contentStrut.StatusCode = res.Status()
			contentStrut.Body = string(res.Body())
		}

		var errLog error
		if errLog = transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
			MerchantCode:    orderX.MerchantCode,
			MerchantOrderNo: orderX.MerchantOrderNo,
			OrderNo:         orderX.OrderNo,
			LogType:         constants.ERROR_MSG,
			LogSource:       constants.API_XF,
			TxOrderSource:   orderX.Source,
			TxOrderType:     orderX.Type,
			Content:         contentStrut,
			TraceId:         trace.SpanContextFromContext(ctx).TraceID().String(),
		}); errLog != nil {
			logx.WithContext(ctx).Errorf("写入交易日志错误:%s", errLog)
		}
	}

	if err != nil {
		logx.WithContext(ctx).Errorf("下发 call NotifyUrl error:%s", err.Error())
		return errorz.New(response.MERCHANT_CALLBACK_ERROR, err.Error())
	} else if res.Status() != 200 {
		logx.WithContext(ctx).Errorf("下发 call NotifyUrl httpStatus:%d, res:%s", res.Status(), string(res.Body()[:]))
		return errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call NotifyUrl httpStatus:%d", res.Status()))
	}
	// 商戶回覆確認
	respString := string(res.Body()[:])
	logx.WithContext(ctx).Infof("UI下发回調商戶，商戶返回 : %v", respString)
	if respString == "success" {
		orderX.IsMerchantCallback = constants.IS_MERCHANT_CALLBACK_YES
		if err1 := db.Table("tx_orders").Updates(orderX).Error; err1 != nil {
			logx.WithContext(ctx).Errorf("UI下发回調成功,但更改回調狀態失敗")
			return errorz.New(response.MERCHANT_CALLBACK_ERROR, err1.Error())
		}
	} else {
		logx.WithContext(ctx).Errorf("錯誤: UI下发回調商戶失敗，err : %v， res : %s", err, res.String())
	}

	return nil
}

func CallChannel_WithdrawOrder(context *context.Context, config *config.Config, order *types.OrderX, orderchannel *types.OrderChannelsX, channel *types.ChannelData) (*vo.ProxyPayRespVO, error) {
	span := trace.SpanFromContext(*context)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              orderchannel.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    fmt.Sprintf("%f", orderchannel.OrderAmount),
		ReceiptAccountNumber: order.MerchantBankNo,
		ReceiptAccountName:   order.MerchantAccountName,
		ReceiptCardProvince:  order.MerchantBankProvince,
		ReceiptCardCity:      order.MerchantBankCity,
		ReceiptCardArea:      "",
		ReceiptCardBranch:    "",
		ReceiptCardBankCode:  order.MerchantBankNo,
		ReceiptCardBankName:  order.MerchantBankName,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	url := fmt.Sprintf("%s:%s/api/proxy-pay", config.Server, channel.ChannelPort)
	//url := "http://154.222.0.115:19081/api/proxy-pay"
	chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("response Status:", chnResp.Status())
		logx.Info("response Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %+v", chnErr.Error(), chnResp)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, chnErr.Error())
	} else if chnResp.Status() != 200 {
		logx.Errorf("渠道返回不正确: %d", chnResp.Status())
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("%d", chnResp.Status()))
	} else if decodeErr := chnResp.DecodeJSON(proxyPayRespVO); decodeErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", decodeErr.Error(), decodeErr)
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, decodeErr.Error())
	} else if proxyPayRespVO.Code != "0" {
		return proxyPayRespVO, errorz.New(proxyPayRespVO.Code, proxyPayRespVO.Message)
	} else if proxyPayRespVO.Data.ChannelOrderNo == "" {
		return proxyPayRespVO, errorz.New(response.INVALID_CHANNEL_ORDER_NO, "ChannelOrderNo: "+proxyPayRespVO.Data.ChannelOrderNo)
	}

	logx.Infof("proxyPayRespVO : %#v", proxyPayRespVO)
	return proxyPayRespVO, nil
}
