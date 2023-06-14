package transactionLogService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"encoding/json"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"net/url"
	"time"
)

//交易日志新增Func
func CreateTransactionLog2(db *gorm.DB, data *types.TransactionLogData) (err error) {

	txLog := types.TxLog{
		MerchantCode:    data.MerchantCode,
		MerchantOrderNo: data.MerchantOrderNo,
		OrderNo:         data.OrderNo,
		LogType:         data.LogType,
		LogSource:       data.LogSource,
		//Content:         data.Content,
		CreatedAt: time.Now().UTC().String(),
	}

	if err = db.Table("tx_log").Create(&txLog).Error; err != nil {
		return
	}

	return nil
}

//交易日志新增Func
func CreateTransactionLog(db *gorm.DB, data *types.TransactionLogData) (err error) {

	var logSource string

	if data.LogSource == "" {
		if data.TxOrderSource == constants.ORDER_SOURCE_BY_PLATFORM {
			if data.TxOrderType == constants.ORDER_TYPE_DF {
				logSource = constants.PLATEFORM_DF
			} else if data.TxOrderType == constants.ORDER_TYPE_NC {
				logSource = constants.PLATEFORM_NC
			} else if data.TxOrderType == constants.ORDER_TYPE_XF {
				logSource = constants.PLATEFORM_XF
			}
		} else if data.TxOrderSource == constants.ORDER_SOURCE_BY_API {
			if data.TxOrderType == constants.ORDER_TYPE_DF {
				logSource = constants.API_DF
			} else if data.TxOrderType == constants.ORDER_TYPE_ZF {
				logSource = constants.API_ZF
			} else if data.TxOrderType == constants.ORDER_TYPE_XF {
				logSource = constants.API_XF
			}
		}
	} else {
		logSource = data.LogSource
	}

	jsonContent, err := json.Marshal(data.Content)
	if err != nil {
		logx.Errorf("產生交易日志錯誤:%s", err.Error())
	}

	txLog := types.TxLog{
		MerchantCode:    data.MerchantCode,
		OrderNo:         data.OrderNo,
		MerchantOrderNo: data.MerchantOrderNo,
		ChannelOrderNo:  data.ChannelOrderNo,
		LogType:         data.LogType,
		LogSource:       logSource,
		Content:         string(jsonContent),
		Log:             produceLogFromTemplate(logSource, data, string(jsonContent)),
		CreatedAt:       time.Now().UTC().String(),
		ErrorCode:       data.ErrCode,
		ErrorMsg:        data.ErrMsg,
		TraceId:         data.TraceId,
	}
	go func() {
		if err = db.Table("tx_log").Create(&txLog).Error; err != nil {
			return
		}
	}()

	return nil
}

// logSource
//	PLATEFORM_NC = "1"
//	API_ZF       = "2"
//	API_DF       = "3"
//	PLATEFORM_DF = "4"
//	API_XF       = "5"
//	PLATEFORM_XF = "6"
func produceLogFromTemplate(logSource string, data *types.TransactionLogData, jsonStr string) (log string) {

	//LogType
	//ERROR_MSG                 = "1" //1:錯誤訊息
	//MERCHANT_REQUEST          = "2" //2:商户请求
	//ERROR_REPLIED_TO_MERCHANT = "3" //3:返回商户错误
	//DATA_REQUEST_CHANNEL      = "4" //4.打给渠道资料
	//RESPONSE_FROM_CHANNEL     = "5" //5.渠道返回资料
	//CALLBACK_FROM_CHANNEL     = "6" //6.渠道回调资料
	//CALLBACK_TO_MERCHANT      = "7" //7.回调给商户
	//RESPONSE_FROM_MERCHANT    = "8" //8.商戶返回訊息

	var err error
	if logSource == constants.PLATEFORM_NC { //V
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req types.OrderInternalCreateRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_NC_MERCHANT_REQUEST, req.MerchantAccountName, req.ChannelAccountName, req.MerchantBankAccount, req.ChannelAccountName, req.MerchantBankName, req.ChannelBankAccount, req.OrderAmount)
		}

	} else if logSource == constants.PLATEFORM_DF {
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req transaction.ProxyOrderUI
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_DF_UI_REQUEST, req.CurrencyCode, fmt.Sprintf("%.2f", req.OrderAmount))

		case constants.CALLBACK_FROM_CHANNEL:
			var req types.PayCallBackRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf("", req.OrderAmount, req.OrderStatus)

		case constants.CALLBACK_TO_MERCHANT:
			var req vo.PayCallBackVO
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_DF_MERCHANT_CALL_BACK, req.OrderAmount, req.OrderStatus)

		case constants.RESPONSE_FROM_MERCHANT:

		}

		if err != nil {
			logx.Errorf("產生交易日誌模板錯誤:", err.Error())
		}

	} else if logSource == constants.PLATEFORM_XF {
		switch data.LogType {
		case constants.MERCHANT_REQUEST:
			var req types.WithdrawApiOrderRequest
			var req2 transaction.WithdrawOrderRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			err = json.Unmarshal([]byte(jsonStr), &req2)
			if len(req.WithdrawName) > 0 {
				log = fmt.Sprintf(constants.PATTERN_XF_TO_PROXY_UI_REQUEST, req.WithdrawName, req.BankName, req.AccountNo, req.OrderAmount)
			} else {
				log = fmt.Sprintf(constants.PATTERN_XF_UI_REQUEST, req2.MerchantAccountName, req2.MerchantBankName, req2.MerchantBankeAccount, fmt.Sprintf("%.2f", req2.OrderAmount))
			}
		case constants.CALLBACK_TO_MERCHANT:
			var req types.WithdrawApiOrderRequest
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", data.Content)), &req)
			log = fmt.Sprintf("", req.NotifyUrl, req.WithdrawName, req.BankName, req.AccountNo, req.OrderAmount)
		}

	} else if logSource == constants.API_DF {
		switch data.LogType {
		case constants.CALLBACK_TO_MERCHANT:
			var req url.Values

			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)

			log = fmt.Sprintf(constants.PATTERN_DF_MERCHANT_CALL_BACK, req.Get("orderAmount"), req.Get("orderStatus"))
		case constants.ERROR_MSG:
			contentStrut := struct {
				StatusCode int
				Body       string
				ErrorMsg   string
			}{}
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &contentStrut)
			log = fmt.Sprintf(constants.PATTERN_ERROR_MSG, contentStrut.StatusCode, contentStrut.ErrorMsg, contentStrut.Body)
		}
	} else if logSource == constants.API_XF {
		switch data.LogType {
		case constants.CALLBACK_TO_MERCHANT:
			var req vo.WithdrawCallBackVO
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &req)
			log = fmt.Sprintf(constants.PATTERN_XF_MERCHANT_CALL_BACK, req.OrderAmount, req.Fee, req.OrderStatus)

		case constants.ERROR_MSG:
			contentStrut := struct {
				StatusCode int
				Body       string
				ErrorMsg   string
			}{}
			err = json.Unmarshal([]byte(fmt.Sprintf("%s", jsonStr)), &contentStrut)
			log = fmt.Sprintf(constants.PATTERN_ERROR_MSG, contentStrut.StatusCode, contentStrut.ErrorMsg, contentStrut.Body)
		}
	}
	return
}
