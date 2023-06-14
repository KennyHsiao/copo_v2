package ordersService

import (
	"com.copo/bo_service/boadmin/internal/config"
	"com.copo/bo_service/boadmin/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/bo"
	"com.copo/bo_service/common/apimodel/bo/searchBO"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"regexp"
)

// 商户请求参数、處理代付訂單、商户费率验证
func ProxyOrder(db *gorm.DB, req *types.ProxyPayRequestX) (rate *types.CorrespondMerChnRate, err error) {
	merchant := &types.Merchant{}
	db.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(merchant)

	// 1. 檢查商户提交參數相關
	err = validateProxyParam(db, req, merchant)
	if err != nil {
		logx.Errorf("商戶%s,代付提單參數錯誤:%s", merchant.Code, i18n.Sprintf(err.Error()))
		return nil, err
	}

	//TODO 代付轉下發

	// 2. 取得商戶配置的費率，以及费率相关验证
	//TODO 儲存DB do and wait for lock 15s
	rate, err = checkProxyOrderAndRate(db, merchant, req)
	if err != nil {
		logx.Errorf("代付提单储存失败%s:", err.Error())
		return rate, err
	}

	return rate, nil
}

/*
	return:
		1. rate      将该订单指定得费率物件返回上一层
        2. err       將错误返回
*/
func checkProxyOrderAndRate(db *gorm.DB, merchant *types.Merchant, req *types.ProxyPayRequestX) (rate *types.CorrespondMerChnRate, err error) {
	//respDTO := &vo.ProxyPayOrderRespVO{}
	//返回物件ProxyPayOrderRespVO
	//userAccount := "TEST0001"
	//var balanceType string
	//var orderFeeProfits []types.OrderFeeProfit

	//检查商户ＡＰＩ提单，代付订单资料是否正确
	err = validProxyPayOrderDataByApi(db, req)
	if err != nil {
		logx.Errorf("检查商户API提单，代付订单资料。%s:%s", err.Error(), i18n.Sprintf(err.Error()))
		return nil, err
	}

	//1. 取得商户对应的代付渠道资料及费率(先收) 计算手续费
	rate1, err1 := merchantchannelrateservice.GetDesignationMerChnRate(db, req.MerchantId, constants.CHN_PAY_TYPE_PROXY_PAY, req.Currency, req.PayTypeSubNo, merchant.BillLadingType)
	logx.Infof("渠道资讯及费率的计算类型:%#v", rate1)
	if err1 != nil {
		logx.Error("商户费率错误。%s:%s", err1.Error(), i18n.Sprintf(err1.Error()))
		logx.Error("商户模式(提單類型 (0=單指、1=多指)): ", merchant.BillLadingType, "提单交易类型:", constants.CHN_PAY_TYPE_PROXY_PAY)
	}
	if rate1 == nil {
		//未配置渠道，列为失败订单
		logx.Error("商户：{}，提单号：{} 未配置渠道，CorrespondMerChnRate={}", merchant.Code, req.OrderNo, rate1)
		jsonData, errParse := json.Marshal(&rate1)
		return nil, errorz.New(response.MERCHANT_IS_NOT_SETTING_CHANNEL, string(jsonData), errParse.Error())
	} else {
		// 判断提单金额最低金额、最高金额
		if req.OrderAmount < rate1.SingleMinCharge {
			logx.Errorf("付款人:%s,银行账号:%s,%f单笔小于最低代付金额%f", req.DefrayName, req.BankNo, req.OrderAmount, rate1.SingleMinCharge)
			return rate1, errorz.New(response.IS_LESS_MINIMUN_AMOUNT, fmt.Sprintf("%f", req.OrderAmount), fmt.Sprintf("%f", rate1.SingleMinCharge))
		} else if req.OrderAmount > rate1.SingleMaxCharge {
			logx.Errorf("付款人:%s,银行账号:%s,%f单笔大于最高代付金额%f", req.DefrayName, req.BankNo, req.OrderAmount, rate1.SingleMaxCharge)
			return rate1, errorz.New(response.IS_GREATER_MXNIMUN_AMOUNT, fmt.Sprintf("%f", req.OrderAmount), fmt.Sprintf("%f", rate1.SingleMinCharge))
		}

		//  代理-取得商户费率层级编号(需提供merchantCoding、agentLayerNo、payTypeCoding)
		// 2. 补当渠道成本增加时，如果尚未重新配置商户费率，商户费率小于成本时，会退回提单。
		channelPayType := &types.ChannelPayType{}
		db.Table("ch_channel_pay_types").Where("code = ?", rate1.ChannelPayTypesCode).Take(channelPayType)
		if rate1.Fee != 0 {
			if rate1.Fee < channelPayType.Fee {
				logx.Errorf("代付提单：%s，商户:%s，代付费率:%f 不可小于渠道成本费率%f ", req.OrderNo, req.MerchantId, rate1.Fee, channelPayType.Fee)
				return rate1, errorz.New(response.RATE_SETTING_ERROR, "代付提单：%s，商户:%s，代付费率:%s 不可小于渠道成本费率%s ", req.OrderNo, req.MerchantId, fmt.Sprintf("%f", rate1.Fee), fmt.Sprintf("%f", channelPayType.Fee))
			}
		}
	}

	return rate1, nil
}

func WithdrawToProxyCallChannel(context *context.Context, config *config.Config, order *types.OrderX, channel *types.ChannelData) (*vo.ProxyPayRespVO, error) {
	span := trace.SpanFromContext(*context)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              order.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    fmt.Sprintf("%f", order.OrderAmount),
		ReceiptAccountNumber: order.MerchantBankAccount,
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

	url := fmt.Sprintf("http://%s:%s/api/proxy-pay", config.ChannelHost, channel.ChannelPort)
	chnResp, chnErr := gozzle.Post(url).Timeout(25).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("response Status:", chnResp.Status())
		logx.Info("response Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
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

/*
	@param respOrder : 代付儲存成功的訂單
    @param rate		 : 商戶配置的費率

	@return error    : call 渠道返回錯誤
*/
func CallChannel(context *context.Context, config *config.Config, merReq *types.ProxyPayRequestX, respOrder *types.OrderX, rate *types.CorrespondMerChnRate) (*vo.ProxyPayRespVO, error) {
	span := trace.SpanFromContext(*context)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              respOrder.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    fmt.Sprintf("%f", respOrder.OrderAmount),
		ReceiptAccountNumber: respOrder.MerchantBankAccount,
		ReceiptAccountName:   respOrder.MerchantAccountName,
		ReceiptCardProvince:  respOrder.MerchantBankProvince,
		ReceiptCardCity:      respOrder.MerchantBankCity,
		ReceiptCardArea:      "",
		ReceiptCardBranch:    merReq.BranchName,
		ReceiptCardBankCode:  respOrder.MerchantBankNo,
		ReceiptCardBankName:  respOrder.MerchantBankName,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	url := fmt.Sprintf("http://%s:%s/api/proxy-pay", config.ChannelHost, rate.ChannelPort)
	chnResp, chnErr := gozzle.Post(url).Timeout(25).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("response Status:", chnResp.Status())
		logx.Info("response Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
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

/*
	@param respOrder : 代付儲存成功的訂單
    @param rate		 : 商戶配置的費率

	@return error    : call 渠道返回錯誤
*/
func CallChannel_BK_PROXY(context *context.Context, config *config.Config, respOrder *types.OrderX, channelPort string) (*vo.ProxyPayRespVO, error) {
	span := trace.SpanFromContext(*context)

	// 新增请求代付请求app 物件 ProxyPayBO
	ProxyPayBO := bo.ProxyPayBO{
		OrderNo:              respOrder.OrderNo,
		TransactionType:      constants.TRANS_TYPE_PROXY_PAY,
		TransactionAmount:    fmt.Sprintf("%f", respOrder.OrderAmount),
		ReceiptAccountNumber: respOrder.MerchantBankAccount,
		ReceiptAccountName:   respOrder.MerchantAccountName,
		ReceiptCardProvince:  respOrder.MerchantBankProvince,
		ReceiptCardCity:      respOrder.MerchantBankCity,
		ReceiptCardArea:      "",
		//ReceiptCardBranch:    req.BranchName,
		ReceiptCardBankCode: respOrder.MerchantBankNo,
		ReceiptCardBankName: respOrder.MerchantBankName,
	}

	// call 渠道app
	ProxyKey, errk := utils.MicroServiceEncrypt(config.ApiKey.ProxyKey, config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errk.Error())
	}

	url := fmt.Sprintf("http://%s:%s/api/proxy-pay", config.ChannelHost, channelPort)
	chnResp, chnErr := gozzle.Post(url).Timeout(25).Trace(span).Header("authenticationProxykey", ProxyKey).JSON(ProxyPayBO)
	//res, err2 := http.Post(url,"application/json",bytes.NewBuffer(body))
	if chnResp != nil {
		logx.Info("response Status:", chnResp.Status())
		logx.Info("response Body:", string(chnResp.Body()))
	}

	proxyPayRespVO := &vo.ProxyPayRespVO{}

	if chnErr != nil {
		logx.Errorf("渠道返回错误: %s， resp: %#v", chnErr.Error(), chnResp)
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

func getMerAllLayerFeeRateInfo(db *gorm.DB, searchBo *searchBO.MerchantLayerRateDataSearchBO) {

}

func validProxyPayOrderDataByApi(db *gorm.DB, req *types.ProxyPayRequestX) (err error) {
	var orderCnt int64
	//1.检查订单号是否重复
	db.Table("tx_orders").Where("merchant_order_no = ?", req.OrderNo).Count(&orderCnt)
	if orderCnt > 0 {
		return errorz.New(response.REPEAT_ORDER_NO, "Merchant OrderNo: "+req.OrderNo)
	}

	//2.验证币别是否可使用
	isCheck := checkCurrencyCodeByApi(db, req.MerchantId, req.Currency)
	if !isCheck {
		return errorz.New(response.CURRENCY_INCONSISTENT, "currency: "+req.Currency)
	}

	return nil
}

func checkCurrencyCodeByApi(db *gorm.DB, merchantCode string, currency string) bool {
	MerchantCurrencyList := []types.MerchantCurrency{}
	err := db.Table("mc_merchant_currencies").Where("").Find(&MerchantCurrencyList).Error
	if err != nil {

	}
	for _, m := range MerchantCurrencyList {
		if m.CurrencyCode == currency {
			return true
		}
	}
	return false
}

func autoFillBankName(db *gorm.DB, req *types.ProxyPayRequestX) (err error) {
	bank := &types.Bank{}
	if req.BankId == "" {
		return errorz.New(response.BANK_CODE_EMPTY)
	} else {
		if err = db.Table("bk_banks").Where("bank_no", req.BankId).Take(bank).Error; err != nil {
			return errorz.New(response.BANK_CODE_INVALID, err.Error(), req.BankId)
		}
		req.BankName = bank.BankName
		return nil
	}
}

func validateProxyParam(db *gorm.DB, req *types.ProxyPayRequestX, merchant *types.Merchant) (err error) {
	// 檢查簽名
	checkSign := utils.VerifySign(req.Sign, req.ProxyPayOrderRequest, merchant.ScrectKey)
	if !checkSign {
		return errorz.New(response.INVALID_SIGN)
	}
	// 檢查新增USDT 钱包地址判断 协定固定 USDT-TRC20
	if req.Currency == "USDT" {
		if isMatch, _ := regexp.MatchString(constants.REGEXP_WALLET_TRC, req.BankNo); !isMatch {
			return errorz.New(response.INVALID_USDT_WALLET_ADDRESS, "USDT_WALLET_ADDRESS: "+req.BankNo)
		}
	}
	// 商戶:
	// 多指定模式。 指定渠道中的渠道，且payTypeSubNo必填
	// 单指定模式。 走指定渠道(唯一一个)
	if merchant.BillLadingType == "1" && req.PayTypeSubNo == "" {
		return errorz.New(response.INVALID_PAY_TYPE_SUB_NO, "PayTypeSubNo: "+req.PayTypeSubNo)
	}
	//======業務參數驗證==========
	if req.AccessType == "" || req.AccessType != constants.ACCESS_TYPE_PROXY {
		return errorz.New(response.INVALID_ACCESS_TYPE, "AccessType: "+req.AccessType)
	}
	if req.MerchantId == "" {
		return errorz.New(response.INVALID_MERCHANT_CODE, "MerchantId: "+req.MerchantId)
	}
	if req.OrderNo == "" {
		return errorz.New(response.INVALID_ORDER_NO, "OrderNo: "+req.OrderNo)
	}
	//4.验证开户行号(银行代码)(必填)(格式必须都为数字)(长度只能为3码)
	isMatch, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankId)
	if req.BankId == "" || !isMatch || len(req.BankId) != 3 {
		logx.Error("开户行号格式不符: ", req.BankId)
		return errorz.New(response.INVALID_BANK_ID, "BankId: "+req.BankId)
	}
	//5.验证开户行名(必填)
	if req.BankName == "" {
		return errorz.New(response.INVALID_BANK_NAME, "BankName: "+req.BankName)
	}

	//6.验证银行卡号(必填)(必须为数字)(长度必须在13~22码)
	isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankNo)
	if req.BankNo == "" || len(req.BankNo) < 13 || len(req.BankNo) > 22 || !isMatch2 {
		return errorz.New(response.INVALID_BANK_NO, "BankNo: "+req.BankNo)
	}

	//7.验证开户人姓名(必填)
	if req.DefrayName == "" {
		return errorz.New(response.INVALID_DEFRAY_NAME, "DefrayName: "+req.DefrayName)
	}

	//8.验证交易金额(必填)
	if req.OrderAmount <= 0 {
		logx.Error("金额错误", req.OrderAmount)
		return errorz.New(response.INVALID_AMOUNT, "OrderAmount: "+fmt.Sprintln("%d", req.OrderAmount))
	}

	isMatch3, _ := regexp.MatchString(constants.REGEXP_URL, req.NotifyUrl)
	//9.验证回调URL格式
	if req.NotifyUrl == "" || !isMatch3 {
		return errorz.New(response.INVALID_NOTIFY_URL, "NotifyUrl: "+req.NotifyUrl)
	}

	//10.验证语系(目前仅支援简体中文)
	if req.Language == "" || req.Language != constants.LANGUAGE_ZH_CN {
		return errorz.New(response.INVALID_LANGUAGE_CODE, "Language: "+req.Language)
	}

	//11 判斷銀行代號自動填入名稱
	if err = autoFillBankName(db, req); err != nil {
		logx.Error("银行代码错误: ", err.Error())
		return errorz.New(response.INVALID_BANK_ID, "BankID: "+req.BankId)
	}

	return nil
}
