package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelReportTotalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelReportTotalLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelReportTotalLogic {
	return ChannelReportTotalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelReportTotalLogic) ChannelReportTotal(req *types.ChannelReportQueryRequestX) (resp *types.ChannelReportTotalResponse, err error) {
	if resp, err = reportService.InterChannelReportTotal(l.svcCtx.MyDB, req, l.ctx); err != nil {
		return
	}

	return
}
