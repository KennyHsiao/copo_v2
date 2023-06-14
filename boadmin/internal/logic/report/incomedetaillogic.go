package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncomeDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomeDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomeDetailLogic {
	return IncomeDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomeDetailLogic) IncomeDetail(req *types.IncomeDetailRequest) (resp *types.IncomeDetailResponse, err error) {
	//var cstZone, loadErr = time.LoadLocation(req.Location)
	//if loadErr != nil {
	//	return nil, errorz.New(response.INVALID_TIMESTAMP, loadErr.Error())
	//}
	nowTime := time.Now()
	startTime, endTime := reportService.GetQueryTodayTime(nowTime)
	var ncRate, zfRate, dfRate, xfHandlingFee float64

	//resp, err = reportService.GetPerDayIncomeDetail(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
	ncRate, err = reportService.GetPerDayIncomeDetailNc(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
	if err != nil {
		return nil, err
	}
	zfRate, err = reportService.GetPerDayIncomeDetailZf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
	if err != nil {
		return nil, err
	}
	dfRate, err = reportService.GetPerDayIncomeDetailDf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
	if err != nil {
		return nil, err
	}
	xfHandlingFee, err = reportService.GetPerDayIncomeDetailXf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
	if err != nil {
		return nil, err
	}
	resp = &types.IncomeDetailResponse{
		NcRate:        ncRate,
		ZfRate:        zfRate,
		DfRate:        dfRate,
		XfHandlingFee: xfHandlingFee,
	}
	return
}
