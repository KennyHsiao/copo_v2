package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantReportQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantReportQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantReportQueryLogic {
	return MerchantReportQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantReportQueryLogic) MerchantReportQuery(req *types.MerchantReportQueryRequestX) (resp *types.MerchantReportQueryResponse, err error) {
	if resp, err = reportService.InterMerchantReport(l.svcCtx.MyDB, req, l.ctx); err != nil {
		return
	}
	return
}
