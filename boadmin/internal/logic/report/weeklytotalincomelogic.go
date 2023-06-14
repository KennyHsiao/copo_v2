package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"context"
	"sync"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WeeklyTotalIncomeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWeeklyTotalIncomeLogic(ctx context.Context, svcCtx *svc.ServiceContext) WeeklyTotalIncomeLogic {
	return WeeklyTotalIncomeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WeeklyTotalIncomeLogic) WeeklyTotalIncome(req *types.WeeklyTotalIncomeRequest) (resp *types.WeeklyTotalIncomeResponse, err error) {
	var weeklyTotalIncomes []types.WeeklyTotalIncome
	//var cstZone, loadErr = time.LoadLocation(req.Location)
	//if loadErr != nil {
	//	return nil, errorz.New(response.INVALID_TIMESTAMP, loadErr.Error())
	//}
	nowTime := time.Now()
	for i := 0; i < 7; i++ {
		startTime, endTime, showTime := reportService.GetWeeklyDate(i, nowTime)
		var incomeDetail types.IncomeDetailResponse
		//incomeDetail, err = reportService.GetPerDayIncomeDetail(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		var ncRate, zfRate, dfRate, xfHandlingFee float64
		var errNc, errZf, errDf, errXf error
		var wg sync.WaitGroup
		wg.Add(4)
		go func() {
			defer wg.Done()
			ncRate, errNc = reportService.GetPerDayIncomeDetailNc(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			zfRate, errZf = reportService.GetPerDayIncomeDetailZf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			dfRate, errDf = reportService.GetPerDayIncomeDetailDf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			xfHandlingFee, errXf = reportService.GetPerDayIncomeDetailXf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		}()

		wg.Wait()

		if errNc != nil {
			return nil, errNc
		}

		if errZf != nil {
			return nil, errZf
		}

		if errDf != nil {
			return nil, errDf
		}

		if errXf != nil {
			return nil, errXf
		}

		incomeDetail.NcRate = ncRate
		incomeDetail.ZfRate = zfRate
		incomeDetail.DfRate = dfRate
		incomeDetail.XfHandlingFee = xfHandlingFee

		weeklyTotalIncome := types.WeeklyTotalIncome{
			Date:         showTime,
			IncomeDetail: incomeDetail,
		}
		weeklyTotalIncomes = append(weeklyTotalIncomes, weeklyTotalIncome)
	}

	resp = &types.WeeklyTotalIncomeResponse{
		List: weeklyTotalIncomes,
	}

	return
}
