package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppropriationCheckBillQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppropriationCheckBillQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) AppropriationCheckBillQueryLogic {
	return AppropriationCheckBillQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppropriationCheckBillQueryLogic) AppropriationCheckBillQuery(req *types.AppropriationCheckBillQueryRequest) (resp *types.AppropriationCheckBillQueryResponse, err error) {
	resp, err = reportService.AppropriationCheckBill(l.svcCtx.MyDB, req, l.ctx)

	return
}
