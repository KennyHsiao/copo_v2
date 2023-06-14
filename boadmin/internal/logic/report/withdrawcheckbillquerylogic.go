package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawCheckBillQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type WithdrawSuccess struct {
	MerchantCode       string
	TotalSuccessAmount float64
	TotalHandlingFee   float64
}

func NewWithdrawCheckBillQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawCheckBillQueryLogic {
	return WithdrawCheckBillQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawCheckBillQueryLogic) WithdrawCheckBillQuery(req *types.WithdrawCheckBillQueryRequestX) (resp *types.WithdrawCheckBillQueryResponse, err error) {
	resp, err = reportService.WithdrawCheckBill(l.svcCtx.MyDB, req, l.ctx)

	return
}
