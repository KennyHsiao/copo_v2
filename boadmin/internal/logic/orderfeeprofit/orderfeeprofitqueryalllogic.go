package orderfeeprofit

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderFeeProfitQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderFeeProfitQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderFeeProfitQueryAllLogic {
	return OrderFeeProfitQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderFeeProfitQueryAllLogic) OrderFeeProfitQueryAll(req types.OrderFeeProfitQueryAllRequest) (resp *types.OrderFeeProfitQueryAllResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
