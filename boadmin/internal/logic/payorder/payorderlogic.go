package payorder

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantchannelrateservice"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/bo"
	"com.copo/bo_service/common/utils"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/gioco-play/gozzle"
	"github.com/jinzhu/copier"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/language"

	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"

	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type PayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderLogic {
	return PayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderLogic) PayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

	if resp, err = l.DoPayOrder(req); err != nil {
		return
	}

	return
}

func (l *PayOrderLogic) DoPayOrder(req types.PayOrderRequestX) (resp *types.PayOrderResponse, err error) {

	var merchant *types.Merchant
	var correspondMerChnRate *types.CorrespondMerChnRate

	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", req.MerchantId).
		Where("status = ?", "1").
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 檢查白名單
	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.IP_DENIED, "IP: "+req.MyIp)
	}

	// 檢查驗簽
	if isSameSign := utils.VerifySign(req.Sign, req.PayOrderRequest, merchant.ScrectKey); !isSameSign {
		return nil, errorz.New(response.SIGN_KEY_FAIL)
	}

	// 检查请求参数
	if err = l.VerifyParam(req, merchant); err != nil {
		return nil, errorz.New(response.INVALID_PARAMETER)
	}

	// 確認是否返回實名制UI畫面
	if req.JumpType == "UI" {
		return l.RequireUserIdPage(req)
	}

	// 取得支付渠道資訊
	if correspondMerChnRate, err = merchantchannelrateservice.GetDesignationMerChnRate(l.svcCtx.MyDB, req.MerchantId, req.PayType, req.Currency, req.PayTypeNo, merchant.BillLadingType); err != nil {
		return
	}

	if correspondMerChnRate.Fee < correspondMerChnRate.ChFee ||
		correspondMerChnRate.HandlingFee < correspondMerChnRate.ChHandlingFee {
		return nil, errorz.New(response.RATE_SETTING_ERROR)
	}

	// 確認商戶訂單號重複
	if isExist, err := model.NewOrder(l.svcCtx.MyDB).IsExistByMerchantOrderNo(merchant.Code, req.OrderNo); isExist {
		return nil, errorz.New(response.ORDER_NUMBER_EXIST, "")
	} else if err != nil {
		return nil, errorz.New(response.SYSTEM_ERROR, err.Error())
	}

	// 確認支付金額上下限
	if err = l.amountVerify(req.OrderAmount, correspondMerChnRate); err != nil {
		return
	}

	// 產生訂單號
	orderNo := model.GenerateOrderNo("ZF")

	// 組成請求json
	payBO := bo.PayBO{
		OrderNo:           orderNo,
		PayType:           correspondMerChnRate.MapCode,
		TransactionAmount: req.OrderAmount,
		BankCode:          req.BankCode,
		PageUrl:           req.PageUrl,
		OrderName:         req.OrderName,
		MerchantId:        req.MerchantId,
		Currency:          req.Currency,
		SourceIp:          req.UserIp,
		UserId:            req.UserId,
		JumpType:          req.JumpType,
	}

	// call 渠道app
	span := trace.SpanFromContext(l.ctx)
	payKey, errk := utils.MicroServiceEncrypt(l.svcCtx.Config.ApiKey.PayKey, l.svcCtx.Config.ApiKey.PublicKey)
	if errk != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	url := correspondMerChnRate.ApiUrl + "/api/pay"
	res, errx := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationPaykey", payKey).JSON(payBO)
	if errx != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	} else if res.Status() != 200 {
		return nil, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call channelApp httpStatus:%d", res.Status()))
	}

	// 處理res
	channelRespBodyVO := vo.PayReplBodyVO{}
	if err = res.DecodeJSON(&channelRespBodyVO); err != nil {
		return nil, errorz.New(response.CHANNEL_REPLY_ERROR, err.Error())
	}
	if channelRespBodyVO.Code != "0" {
		return nil, errorz.New(channelRespBodyVO.Code, channelRespBodyVO.Message)
	}
	payReplyVO := channelRespBodyVO.Data

	var rpcPayOrder transaction.PayOrder
	var rpcRate transaction.CorrespondMerChnRate
	copier.Copy(&rpcPayOrder, &req)
	copier.Copy(&rpcRate, correspondMerChnRate)
	// CALL transactionc PayOrderTranaction
	if _, err = l.svcCtx.TransactionRpc.PayOrderTranaction(l.ctx, &transaction.PayOrderRequest{
		PayOrder:       &rpcPayOrder,
		Rate:           &rpcRate,
		OrderNo:        orderNo,
		ChannelOrderNo: payReplyVO.ChannelOrderNo,
	}); err != nil {
		return
	}

	// 判斷返回格式 1.html, 2.json  3.url
	resp = l.getResp(req, payReplyVO, orderNo)
	i18n.SetLang(language.English)
	resp.RespCode = response.API_SUCCESS
	resp.RespMsg = i18n.Sprintf(response.API_SUCCESS)
	resp.Status = 0
	resp.Sign = utils.SortAndSign2(*resp, merchant.ScrectKey)

	return
}

func (l *PayOrderLogic) RequireUserIdPage(req types.PayOrderRequestX) (*types.PayOrderResponse, error) {
	var orderNo string
	var url string

	orderNo = model.GenerateOrderNo("ZF")

	// TODO: 暫存資料至 redis

	// TODO: 組成url
	url = "http://154.222.0.111/#/checkoutPlayer" + "?id=" + orderNo + "&lang=zh-CN"

	return &types.PayOrderResponse{
		Status:     0,
		PayOrderNo: orderNo,
		BankCode:   req.BankCode,
		Type:       "url",
		Info:       url,
	}, nil
}

func (l *PayOrderLogic) VerifyParam(req types.PayOrderRequestX, merchant *types.Merchant) error {
	// 開啟多選商戶 必需給指定代碼
	if merchant.BillLadingType == "1" && len(req.PayTypeNo) == 0 {
		return errorz.New(response.NO_CHANNEL_SET, "")
	}
	// USDT 限制PayType
	if strings.EqualFold(req.Currency, "USDT") && !utils.Contain(req.PayType, []string{"UT", "UE", "UU"}) {
		return errorz.New(response.INVALID_USDT_TYPE, fmt.Sprintf("(payType): %s", req.PayType))
	}

	return nil
}

func (l *PayOrderLogic) amountVerify(orderAmount string, correspondMerChnRate *types.CorrespondMerChnRate) (err error) {

	var amount float64

	if amount, err = strconv.ParseFloat(orderAmount, 64); err != nil {
		return errorz.New(response.INVALID_AMOUNT, fmt.Sprintf("(orderAmount): %s", orderAmount))
	}

	if amount < 0 {
		return errorz.New(response.ORDER_AMOUNT_INVALID, fmt.Sprintf("(orderAmount): %f", amount))
	}
	if amount > correspondMerChnRate.SingleMaxCharge {
		return errorz.New(response.ORDER_AMOUNT_LIMIT_MAX, fmt.Sprintf("(orderAmount): %f", amount))
	}
	if amount < correspondMerChnRate.SingleMinCharge {
		return errorz.New(response.ORDER_AMOUNT_LIMIT_MIN, fmt.Sprintf("(orderAmount): %f", amount))
	}

	return
}

func (l *PayOrderLogic) getResp(req types.PayOrderRequestX, replyVO vo.PayReplyVO, orderNo string) (resp *types.PayOrderResponse) {

	resp = &types.PayOrderResponse{}
	// 預設url
	info := replyVO.PayPageInfo

	// PayPageType 非url 非json 就跑 html
	if !strings.EqualFold(replyVO.PayPageType, "url") && !strings.EqualFold(replyVO.PayPageType, "json") {
		if !strings.EqualFold(replyVO.PayPageType, "html") {
			logx.Error(fmt.Sprintf("Channel Reply Type:%s error", replyVO.PayPageType))
		}
		// TODO: 實作包HTML功能
	}

	if strings.EqualFold(replyVO.PayPageType, "json") && replyVO.IsCheckOutMer {
		// TODO: 實作短網址 存入redis
		info = "shorturl"
	}
	resp.BankCode = req.BankCode
	resp.Info = info
	resp.PayOrderNo = orderNo
	resp.Type = replyVO.PayPageType

	return

}
