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
	"github.com/zeromicro/go-zero/core/logx"
)

type CallBacPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCallBacPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) CallBacPayOrderLogic {
	return CallBacPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CallBacPayOrderLogic) CallBacPayOrder(req types.CallBackPayOrderRequest) error {
	var order *types.OrderX
	var merchant *types.Merchant

	logx.WithContext(l.ctx).Errorf("UI支付回调商户, 单号:%s", req.OrderNo)

	// 取得訂單
	if err := l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", req.OrderNo).Take(&order).Error; err != nil {
		return errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	}

	// 只有成功/失敗單 & 不是不需回調的單才能回調
	if order.Status != constants.SUCCESS && order.Status != constants.FAIL {
		return errorz.New(response.ONLY_SUCCESSFUL_ORDER_CAN_CALL_BACK)
	} else if order.IsMerchantCallback == constants.MERCHANT_CALL_BACK_DONT_USE {
		return errorz.New(response.ORDER_DOES_NOT_NEED_CALL_BACK)
	}

	// 取得商戶
	if err := l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ?", order.MerchantCode).
		Take(&merchant).Error; err != nil {
		return err
	}

	payCallBackVO := vo.PayCallBackVO{
		AccessType:   "1",
		Language:     "zh-CN",
		MerchantId:   order.MerchantCode,
		OrderNo:      order.MerchantOrderNo,
		OrderTime:    order.CreatedAt.Format("20060102150405000"),
		PayOrderTime: order.TransAt.Time().Format("20060102150405000"),
		Fee:          fmt.Sprintf("%.2f", order.TransferHandlingFee),
		PayOrderId:   order.OrderNo,
	}

	// 若有實際金額則回覆實際
	if order.ActualAmount > 0 {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", order.ActualAmount)
	} else {
		payCallBackVO.OrderAmount = fmt.Sprintf("%.2f", order.OrderAmount)
	}

	// API 支付状态 0：处理中，1：成功，2：失败，3：成功(人工确认)
	payCallBackVO.OrderStatus = "0"
	if order.Status == constants.FAIL {
		payCallBackVO.OrderStatus = "2"
	} else if order.Status == constants.SUCCESS {
		payCallBackVO.OrderStatus = "1"
	}

	// 加簽
	payCallBackVO.Sign = utils.SortAndSign2(payCallBackVO, merchant.ScrectKey)

	// 回調商戶
	if err := callNoticeUrlService.CallNoticeUrlForZF(l.ctx, l.svcCtx.MyDB, order.NotifyUrl, payCallBackVO); err != nil {
		return err
	}

	return nil
}
