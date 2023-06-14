package proxypayorder

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantbalanceservice"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type ProxyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayOrderLogic {
	return ProxyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayOrderLogic) ProxyPayOrder(merReq *types.ProxyPayRequestX) (resp *types.ProxyPayOrderResponse, errMer error) {
	logx.Info("Enter proxy-order:", merReq)

	// 1. 檢查白名單及商户号
	merchantKey, errWhite := l.CheckMerAndWhiteList(merReq)
	if errWhite != nil {
		logx.Error("商戶號及白名單檢查錯誤: ", errWhite.Error())
		return nil, errWhite
	}

	// 2. 處理商户提交参数、訂單驗證
	rate, errCreate := ordersService.ProxyOrder(l.svcCtx.MyDB, merReq)
	if errCreate != nil {
		logx.Error("代付提單商户提交参数驗證錯誤: ", errCreate.Error())
		return nil, errCreate
	}

	balanceType, errBalance := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, rate.ChannelCode, constants.ORDER_TYPE_DF)
	if errBalance != nil {
		return nil, errBalance
	}

	// 3. 依balanceType决定要呼叫哪种transaction方法
	//    呼叫 transaction rpc (merReq, rate) (ProxyOrderNo) 并产生订单

	//產生rpc 代付需要的請求的資料物件
	ProxyPayOrderRequest, rateRpc := generateRpcdata(&merReq.ProxyPayOrderRequest, rate)

	var errRpc error
	var res *transaction.ProxyOrderResponse
	if balanceType == "DFB" {
		res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTranaction_DFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	} else if balanceType == "XFB" {
		res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTranaction_XFB(l.ctx, &transaction.ProxyOrderRequest{
			Req:  ProxyPayOrderRequest,
			Rate: rateRpc,
		})
	}

	if errRpc != nil {
		logx.Error("代付提單:", errRpc.Error())
		return nil, errorz.New(response.FAIL, errRpc.Error())
	} else {
		logx.Infof("代付交易rpc完成，%s 錢包扣款完成: %#v", balanceType, res)
	}

	var queryErr error
	var respOrder = &types.OrderX{}
	if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
		logx.Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
		return nil, errorz.New(response.FAIL, queryErr.Error())
	}

	// 4: call channel (不論是否有成功打到渠道，都要返回給商戶success，一渠道返回訂單狀態決定此訂單狀態(代處理/處理中))
	var errCHN error
	proxyPayRespVO := &vo.ProxyPayRespVO{}
	proxyPayRespVO, errCHN = ordersService.CallChannel(&l.ctx, &l.svcCtx.Config, merReq, respOrder, rate)

	//5. 返回給商戶物件
	var proxyResp = types.ProxyPayOrderResponse{}
	i18n.SetLang(language.English)
	if errCHN != nil {
		logx.Errorf("代付提單: %s ，渠道返回錯誤: %s, %#v", respOrder.OrderNo, errCHN.Error(), proxyPayRespVO)
		proxyResp.RespCode = response.CHANNEL_REPLY_ERROR
		proxyResp.RespMsg = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Code: " + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message
		respOrder.Status = constants.FAIL
		respOrder.ErrorNote = i18n.Sprintf(response.CHANNEL_REPLY_ERROR) + ": Code: " + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message
		respOrder.ErrorType = "2"
		//TODO 将商户钱包加回
	} else {
		//条整订单状态从"待处理" 到 "处理中"
		respOrder.Status = constants.PROCESSING
		proxyResp.RespCode = response.API_SUCCESS
		proxyResp.RespMsg = i18n.Sprintf(response.API_SUCCESS) //固定回商戶成功
	}

	// 更新订单
	if err := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; err != nil {
		logx.Error("代付订单更新状态错误: ", err.Error())
	}

	// 5. 依渠道返回给予订单状态
	var orderStatus string
	if respOrder.Status == constants.FAIL {
		orderStatus = "2"
	} else {
		orderStatus = "0"
	}

	proxyResp.MerchantId = respOrder.MerchantCode
	proxyResp.OrderNo = respOrder.MerchantOrderNo
	proxyResp.PayOrderNo = respOrder.OrderNo
	proxyResp.OrderStatus = orderStatus //渠道返回成功: "處理中" 失敗: "失敗"
	proxyResp.Sign = utils.SortAndSign2(proxyResp, merchantKey)

	return &proxyResp, nil
}

//检查商户号是否存在以及IP是否为白名单，若无误则返回"商户密鑰"
func (l *ProxyPayOrderLogic) CheckMerAndWhiteList(req *types.ProxyPayRequestX) (merchantKey string, err error) {
	merchant := &types.Merchant{}
	// 檢查白名單
	if err = l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else if err == nil && merchant != nil && merchant.Status != constants.MerchantStatusEnable {
			return "", errorz.New(response.MERCHANT_ACCOUNT_NOT_FOUND, "商户号:"+merchant.Code)
		} else {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if isWhite := merchantsService.IPChecker(req.Ip, merchant.ApiIP); !isWhite {
		return "", errorz.New(response.API_IP_DENIED, "IP: "+req.Ip)
	}
	return merchant.ScrectKey, nil
}

// 產生rpc 代付需要的請求的資料物件
func generateRpcdata(merReq *types.ProxyPayOrderRequest, rate *types.CorrespondMerChnRate) (*transaction.ProxyPayOrderRequest, *transaction.CorrespondMerChnRate) {

	ProxyPayOrderRequest := &transaction.ProxyPayOrderRequest{
		AccessType:   merReq.AccessType,
		MerchantId:   merReq.MerchantId,
		Sign:         merReq.Sign,
		NotifyUrl:    merReq.NotifyUrl,
		Language:     merReq.Language,
		OrderNo:      merReq.OrderNo,
		BankId:       merReq.BankId,
		BankName:     merReq.BankName,
		BankProvince: merReq.BankProvince,
		BankCity:     merReq.BankCity,
		BranchName:   merReq.BranchName,
		BankNo:       merReq.BankNo,
		OrderAmount:  merReq.OrderAmount,
		DefrayName:   merReq.DefrayName,
		DefrayId:     merReq.DefrayId,
		DefrayMobile: merReq.DefrayMobile,
		DefrayEmail:  merReq.DefrayEmail,
		Currency:     merReq.Currency,
		PayTypeSubNo: merReq.PayTypeSubNo,
	}
	rateRpc := &transaction.CorrespondMerChnRate{
		MerchantCode:        rate.MerchantCode,
		ChannelPayTypesCode: rate.ChannelPayTypesCode,
		ChannelCode:         rate.ChannelCode,
		PayTypeCode:         rate.PayTypeCode,
		Designation:         rate.Designation,
		DesignationNo:       rate.DesignationNo,
		Fee:                 rate.Fee,
		HandlingFee:         rate.HandlingFee,
		ChFee:               rate.ChFee,
		ChHandlingFee:       rate.ChHandlingFee,
		SingleMinCharge:     rate.SingleMinCharge,
		SingleMaxCharge:     rate.SingleMaxCharge,
		CurrencyCode:        rate.CurrencyCode,
		ApiUrl:              rate.ApiUrl,
	}

	return ProxyPayOrderRequest, rateRpc
}
