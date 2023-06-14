package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/callNoticeUrlService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ConfirmPayOrderLogic {
	return ConfirmPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmPayOrderLogic) ConfirmPayOrder(req types.ConfirmPayOrderRequest) error {
	var rpcRequest transaction.ConfirmPayOrderRequest
	copier.Copy(&rpcRequest, &req)
	// CALL transactionc
	rpcResp, err2 := l.svcCtx.TransactionRpc.ConfirmPayOrderTransaction(l.ctx, &rpcRequest)
	if err2 != nil {
		return err2
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "PayOrderTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	if len(rpcResp.NotifyUrl) > 0 {
		if err2 := l.callNoticeURL(rpcResp); err2 != nil {
			// 回調失敗不影響確認收款
			logx.Errorf("回調失敗:%#v", err2)
		}
	}

	return nil
}

func (l *ConfirmPayOrderLogic) callNoticeURL(callBackResp *transaction.ConfirmPayOrderResponse) (err error) {
	var merchant *types.Merchant
	// 取得商戶
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", callBackResp.MerchantCode).
		Take(&merchant).Error; err != nil {
		return err
	}

	payCallBackVO := vo.PayCallBackVO{
		AccessType:   "1",
		Language:     "zh-CN",
		MerchantId:   callBackResp.MerchantCode,
		OrderNo:      callBackResp.MerchantOrderNo,
		OrderTime:    callBackResp.OrderTime,
		PayOrderTime: callBackResp.PayOrderTime,
		Fee:          fmt.Sprintf("%.2f", callBackResp.TransferHandlingFee),
		PayOrderId:   callBackResp.OrderNo,
	}

	// 若有實際金額則回覆實際
	if callBackResp.ActualAmount > 0 {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", callBackResp.ActualAmount)
	} else {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", callBackResp.OrderAmount)
	}

	// API 支付状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	payCallBackVO.OrderStatus = "0"
	if callBackResp.Status == constants.SUCCESS {
		payCallBackVO.OrderStatus = "1"
	} else if callBackResp.Status == constants.FAIL {
		payCallBackVO.OrderStatus = "2"
	}

	payCallBackVO.Sign = utils.SortAndSign2(payCallBackVO, merchant.ScrectKey)

	// 回調商戶
	if err2 := callNoticeUrlService.CallNoticeUrlForZF(l.ctx, l.svcCtx.MyDB, callBackResp.NotifyUrl, payCallBackVO); err != nil {
		return err2
	}
	return nil
}
