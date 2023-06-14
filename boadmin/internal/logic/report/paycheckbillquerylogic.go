package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayCheckBillQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayCheckBillQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayCheckBillQueryLogic {
	return PayCheckBillQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayCheckBillQueryLogic) PayCheckBillQuery(req *types.PayCheckBillQueryRequestX) (resp *types.PayCheckBillQueryResponse, err error) {
	resp, err = reportService.PayCheckBill(l.svcCtx.MyDB, req, l.ctx)
	return
}
