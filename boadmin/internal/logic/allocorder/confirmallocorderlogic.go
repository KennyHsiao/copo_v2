package allocorder

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmAllocOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmAllocOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ConfirmAllocOrderLogic {
	return ConfirmAllocOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmAllocOrderLogic) ConfirmAllocOrder(req *types.ConfirmAllocOrderRequest) error {
	var order types.OrderX

	// 確認訂單是否存在
	if err := l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", req.OrderNo).
		Take(&order).Error; err != nil {
		return errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
	}

	// 只有交易中能確認撥款
	if order.Status != constants.TRANSACTION {
		return errorz.New(response.ORDER_STATUS_WRONG)
	}

	// 手動確認收款 實際金額 = 訂單金額
	order.ActualAmount = order.OrderAmount
	order.TransAt = types.JsonTime{}.New()
	order.Status = constants.SUCCESS
	order.Memo = req.Comment + " \n" + order.Memo

	// 編輯訂單
	if err := l.svcCtx.MyDB.Table("tx_orders").Updates(&order).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE)
	}

	return nil
}
