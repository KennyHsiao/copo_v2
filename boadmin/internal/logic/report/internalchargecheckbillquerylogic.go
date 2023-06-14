package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type InternalChargeCheckBillQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInternalChargeCheckBillQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) InternalChargeCheckBillQueryLogic {
	return InternalChargeCheckBillQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InternalChargeCheckBillQueryLogic) InternalChargeCheckBillQuery(req *types.PayCheckBillQueryRequestX) (resp *types.InternalChargeCheckBillQueryResponse, err error) {
	resp, err = reportService.InterChargeCheckBill(l.svcCtx.MyDB, req, l.ctx)

	return
}
