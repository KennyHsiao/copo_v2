package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelReportQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelReportQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelReportQueryLogic {
	return ChannelReportQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelReportQueryLogic) ChannelReportQuery(req *types.ChannelReportQueryRequestX) (resp *types.ChannelReportQueryresponse, err error) {
	if resp, err = reportService.InterChannelReport(l.svcCtx.MyDB, req, l.ctx); err != nil {
		return
	}
	return
}
