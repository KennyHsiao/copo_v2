package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayCheckBillQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayCheckBillQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayCheckBillQueryLogic {
	return ProxyPayCheckBillQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayCheckBillQueryLogic) ProxyPayCheckBillQuery(req *types.PayCheckBillQueryRequestX) (resp *types.ProxyPayCheckBillQueryResponse, err error) {

	resp, err = reportService.ProxyPayCheckBill(l.svcCtx.MyDB, req, l.ctx)

	return
}
