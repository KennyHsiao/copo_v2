package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncomeExpenseQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomeExpenseQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomeExpenseQueryAllLogic {
	return IncomeExpenseQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomeExpenseQueryAllLogic) IncomeExpenseQueryAll(req types.IncomeExpenseQueryRequestX) (resp *types.IncomeExpenseQueryResponseX, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.MerchantCode = jwtMerchantCode
	}
	if resp, err = orderrecordService.IncomeExpenseQueryAll(l.svcCtx.MyDB, req, l.ctx); err != nil {
		return
	}
	return
}
