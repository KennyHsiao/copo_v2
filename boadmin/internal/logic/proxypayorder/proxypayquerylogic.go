package proxypayorder

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayQueryLogic {
	return ProxyPayQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayQueryLogic) ProxyPayQuery(merReq *types.ProxyPayOrderQueryRequestX) (resp *types.ProxyPayOrderQueryResponse, err error) {
	logx.Info("Enter proxy-query:", merReq)
	// 1. 檢查白名單、商户号，單號是否存在
	merchantKey, errWhite := l.CheckMerAndWhiteList(merReq)
	if errWhite != nil {
		logx.Error("商戶號及白名單檢查錯誤: ", errWhite.Error())
		return nil, errWhite
	}

	// 2. call 渠道

	res, errRpc := l.svcCtx.TransactionRpc.ProxyOrderTranaction_XFB(l.ctx, &transaction.ProxyOrderRequest{
		Req:         nil,
		Rate:        nil,
		BalanceType: "",
	})

	if errRpc != nil {
		logx.Errorf("商戶: %s ，%s 代付查單錯誤: %s", merReq.MerchantId, merReq.OrderNo, errRpc.Error())
		return nil, errorz.New(response.FAIL, errRpc.Error())
	} else {
		logx.Infof("代付查單rpc完成。 #%v", res)
	}

	respOrder := &types.Order{}
	//返回給商戶查詢物件
	i18n.SetLang(language.English)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS) //固定回商戶成功
	resp.MerchantId = respOrder.MerchantCode
	resp.OrderNo = respOrder.MerchantOrderNo
	resp.PayOrderNo = respOrder.OrderNo
	resp.OrderStatus = respOrder.Status //
	resp.Sign = utils.SortAndSign2(*resp, merchantKey)

	return resp, nil
}

func validProxyPayOrderDataByApi() {

}

//检查商户号是否存在以及IP是否为白名单，若无误则返回"商户密鑰"
func (l *ProxyPayQueryLogic) CheckMerAndWhiteList(req *types.ProxyPayOrderQueryRequestX) (merchantKey string, err error) {
	merchant := &types.Merchant{}
	// 1.檢查白名單
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

	//2.检查订单号是否重复
	if order, queryErr := model.QueryOrderByOrderNo(l.svcCtx.MyDB, "", req.OrderNo); queryErr != nil && order != nil {
		return "", errorz.New(response.REPEAT_ORDER_NO, "Merchant OrderNo: "+req.OrderNo)
	}

	return merchant.ScrectKey, nil
}
