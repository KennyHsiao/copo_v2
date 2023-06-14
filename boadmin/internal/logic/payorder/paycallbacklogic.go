package payorder

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayCallBackLogic {
	return PayCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayCallBackLogic) PayCallBack(req types.PayCallBackRequest) (resp *types.PayCallBackResponse, err error) {

	// 只能回調成功/失敗
	if req.OrderStatus != "20" && req.OrderStatus != "30" {
		return nil, errorz.New(response.TRANSACTION_FAILURE, fmt.Sprintf("(req OrderStatus): %s", req.OrderStatus))
	}

	// CALL transactionc PayOrderTranaction
	callBackResp, err3 := l.svcCtx.TransactionRpc.PayCallBackTranaction(l.ctx, &transaction.PayCallBackRequest{
		CallbackTime:   req.CallbackTime,
		ChannelOrderNo: req.ChannelOrderNo,
		OrderAmount:    req.OrderAmount,
		OrderStatus:    req.OrderStatus,
		PayOrderNo:     req.PayOrderNo,
	})
	if err3 != nil {
		return nil, err3
	}

	logx.Info("PayCallBackTranaction return:", callBackResp)

	// 只有成功單 且 有回掉網址 才回調
	if req.OrderStatus == "20" && len(callBackResp.NotifyUrl) > 0 {
		l.callNoticeURL(callBackResp)
	}

	return
}

func (l *PayCallBackLogic) callNoticeURL(callBackResp *transaction.PayCallBackResponse) (err error) {

	var merchant *types.Merchant
	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", callBackResp.MerchantCode).
		Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.INVALID_MERCHANT_CODE, err.Error())
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	// 状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	var orderStatus = "0"

	// 訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
	if callBackResp.Status == "20" {
		orderStatus = "2"
	} else if callBackResp.Status == "30" {
		orderStatus = "3"
	}

	payCallBackVO := vo.PayCallBackVO{
		AccessType:   "1",
		Language:     "zh-CN",
		MerchantId:   callBackResp.MerchantCode,
		OrderNo:      callBackResp.MerchantOrderNo,
		OrderAmount:  fmt.Sprintf("%.2f", callBackResp.OrderAmount),
		OrderTime:    callBackResp.OrderTime,
		PayOrderTime: callBackResp.PayOrderTime,
		Fee:          fmt.Sprintf("%.2f", callBackResp.TransferHandlingFee),
		OrderStatus:  orderStatus,
		PayOrderId:   callBackResp.OrderNo,
	}

	payCallBackVO.Sign = utils.SortAndSign2(payCallBackVO, merchant.ScrectKey)

	// TODO: 通知商戶

	return
}
