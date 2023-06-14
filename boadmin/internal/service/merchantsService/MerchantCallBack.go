package merchantsService

import (
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/common/constants"
	"context"
	"net/url"
	"strconv"

	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

/*
	回調-商戶代付結果(須注意Scheduled Project 也有一組再用，修改時要注意)
	@return
*/
func PostCallbackToMerchant(db *gorm.DB, context *context.Context, orderX *types.OrderX) (err error) {
	span := trace.SpanFromContext(*context)
	merchant := &types.Merchant{}
	if err = db.Table("mc_merchants").Where("code = ?", orderX.MerchantCode).Find(merchant).Error; err != nil {
		return
	}
	changeStatus := changeOrderStatusToMerchant(orderX.Status)

	ProxyPayCallBackMerRespVO := url.Values{}
	ProxyPayCallBackMerRespVO.Set("merchantId", orderX.MerchantCode)
	ProxyPayCallBackMerRespVO.Set("orderNo", orderX.MerchantOrderNo)
	ProxyPayCallBackMerRespVO.Set("payOrderNo", orderX.OrderNo)
	ProxyPayCallBackMerRespVO.Set("orderStatus", changeStatus)
	ProxyPayCallBackMerRespVO.Set("orderAmount", strconv.FormatFloat(orderX.OrderAmount, 'f', 2, 64))
	ProxyPayCallBackMerRespVO.Set("fee", strconv.FormatFloat(orderX.Fee, 'f', 2, 64))
	ProxyPayCallBackMerRespVO.Set("payOrderTime", orderX.TransAt.Time().Format("200601021504"))
	if orderX.CurrencyCode == "PHP" {
		ProxyPayCallBackMerRespVO.Set("errorNote", orderX.ErrorNote)
	}

	sign := utils.SortAndSignFromUrlValues(ProxyPayCallBackMerRespVO, merchant.ScrectKey)
	ProxyPayCallBackMerRespVO.Set("sign", sign)
	logx.WithContext(*context).Infof("代付提单 %s ，回调商户URL= %s，回调资讯= %#v", orderX.OrderNo, orderX.NotifyUrl, ProxyPayCallBackMerRespVO)

	// 写入交易日志
	if err := transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
		MerchantCode:    orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		ChannelOrderNo:  orderX.ChannelOrderNo,
		LogType:         constants.CALLBACK_TO_MERCHANT,
		//LogSource: constants.API_DF,
		TxOrderSource: orderX.Source,
		TxOrderType:   orderX.Type,
		Content:       ProxyPayCallBackMerRespVO,
		TraceId:       trace.SpanContextFromContext(*context).TraceID().String(),
	}); err != nil {
		logx.WithContext(*context).Errorf("写入交易日志错误:%s", err)
	}

	//TODO retry post for 10 times and 2s between each reqeuest
	//TODO 內部測試，測完需移除
	//merResp, merCallBackErr := gozzle.Post("http://172.16.204.115:8083/dior/merchant-api/merchant-call-back").Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)
	logx.WithContext(*context).Infof("代付提单 %s，回调网址:%s，UI回调商户請求參數 %#v", ProxyPayCallBackMerRespVO.Get("payOrderNo"), orderX.NotifyUrl, ProxyPayCallBackMerRespVO)
	merResp, merCallBackErr := gozzle.Post(orderX.NotifyUrl).Timeout(10).Trace(span).Form(ProxyPayCallBackMerRespVO)

	if merResp != nil {
		// 写入交易日志
		if err := transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
			MerchantCode:    orderX.MerchantCode,
			MerchantOrderNo: orderX.MerchantOrderNo,
			OrderNo:         orderX.OrderNo,
			ChannelOrderNo:  orderX.ChannelOrderNo,
			LogType:         constants.RESPONSE_FROM_MERCHANT,
			//LogSource: constants.API_DF,
			TxOrderSource: orderX.Source,
			TxOrderType:   orderX.Type,
			Content:       merResp,
			TraceId:       trace.SpanContextFromContext(*context).TraceID().String(),
		}); err != nil {
			logx.WithContext(*context).Errorf("写入交易日志错误:%s", err)
		}
	}

	if merCallBackErr != nil || merResp.Status() != 200 {
		statusCode := 0
		body := ""
		errorMsg := ""
		if merResp != nil {
			statusCode = merResp.Status()
			body = string(merResp.Body())
		}
		if merCallBackErr != nil {
			errorMsg = merCallBackErr.Error()
		}

		contentStrut := struct {
			StatusCode int
			Body       string
			ErrorMsg   string
		}{
			StatusCode: statusCode,
			Body:       body,
			ErrorMsg:   errorMsg,
		}

		var errLog error
		if errLog = transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
			MerchantCode:    orderX.MerchantCode,
			MerchantOrderNo: orderX.MerchantOrderNo,
			OrderNo:         orderX.OrderNo,
			LogType:         constants.ERROR_MSG,
			//LogSource: constants.API_DF,
			TxOrderSource: orderX.Source,
			TxOrderType:   orderX.Type,
			Content:       contentStrut,
			TraceId:       trace.SpanContextFromContext(*context).TraceID().String(),
		}); errLog != nil {
			logx.WithContext(*context).Errorf("写入交易日志错误:%s", errLog)
		}

		if merCallBackErr != nil {
			logx.WithContext(*context).Errorf("代付提单%s UI回调商户异常，錯誤: %#v", ProxyPayCallBackMerRespVO.Get("payOrderNo"), merCallBackErr.Error())
			return errorz.New(response.MERCHANT_CALLBACK_ERROR, merCallBackErr.Error())
		} else if merResp.Status() != 200 {
			logx.WithContext(*context).Errorf("UI代付回调，响应状态 %d 错误", merResp.Status())
			return errorz.New(response.MERCHANT_CALLBACK_ERROR)
		}
	}
	logx.WithContext(*context).Infof("代付提单 %s，回调网址:%s,UI回调商户請求參數 %#v，商戶返回: %#v", ProxyPayCallBackMerRespVO.Get("payOrderNo"), orderX.NotifyUrl, ProxyPayCallBackMerRespVO, string(merResp.Body()))

	return
}

func changeOrderStatusToMerchant(status string) string {
	var changeStatus string

	if status == "0" {
		changeStatus = "0"
	} else if status == "1" { //(0:待處理 1:處理中 2:交易中 20:成功 30:失敗 31:凍結)
		changeStatus = "4"
	} else if status == "2" {
		changeStatus = "3"
	} else if status == "20" {
		changeStatus = "1"
	} else if status == "30" || status == "31" {
		changeStatus = "2"
	}

	return changeStatus
}
