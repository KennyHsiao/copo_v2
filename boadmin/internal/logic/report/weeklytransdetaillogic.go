package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/common/constants"
	"context"
	"sync"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WeeklyTransDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWeeklyTransDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) WeeklyTransDetailLogic {
	return WeeklyTransDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WeeklyTransDetailLogic) WeeklyTransDetail(req *types.WeeklyTransDetailRequest) (resp *types.WeeklyTransDetailResponse, err error) {
	var weeklyTransDetails []types.WeeklyTransDetail
	//var cstZone, loadErr = time.LoadLocation(req.Location)
	//if loadErr != nil {
	//	return nil, errorz.New(response.INVALID_TIMESTAMP, loadErr.Error())
	//}
	nowTime := time.Now()
	//startTime, endTime, _ := reportService.GetWeeklyDate(7, nowTime)
	//weeklyTransDetails, errAll := reportService.GetWeeklyOrderAmount(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime)
	//if errAll != nil {
	//	return nil, errAll
	//}
	//
	//if len(weeklyTransDetails) > 0 {
	//	for i, detail := range weeklyTransDetails {
	//		ds := strings.Split(detail.Date, "T")
	//		weeklyTransDetails[i].Date = ds[0]
	//	}
	//} else {
	//
	//}
	//
	//resp = &types.WeeklyTransDetailResponse{
	//	List: weeklyTransDetails,
	//}

	for i := 0; i < 7; i++ {
		startTime, endTime, showTime := reportService.GetWeeklyDate(i, nowTime)

		var wg sync.WaitGroup
		wg.Add(4)
		var ncOrderAmount, zfOrderAmount, dfOrderAmount, xfOrderAmount float64
		var err1, err2, err3, err4 error
		go func() {
			defer wg.Done()
			ncOrderAmount, err1 = reportService.GetOneOrderTypeTotalAmount(l.svcCtx.MyDB, constants.ORDER_TYPE_NC, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			zfOrderAmount, err2 = reportService.GetOneOrderTypeTotalAmount(l.svcCtx.MyDB, constants.ORDER_TYPE_ZF, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			dfOrderAmount, err3 = reportService.GetOneOrderTypeTotalAmount(l.svcCtx.MyDB, constants.ORDER_TYPE_DF, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		go func() {
			defer wg.Done()
			xfOrderAmount, err4 = reportService.GetOneOrderTypeTotalAmount(l.svcCtx.MyDB, constants.ORDER_TYPE_XF, req.CurrencyCode, startTime, endTime, l.ctx)
		}()
		wg.Wait()

		if err1 != nil {
			return nil, err1
		}

		if err2 != nil {
			return nil, err2
		}

		if err3 != nil {
			return nil, err3
		}

		if err4 != nil {
			return nil, err4
		}

		weeklyTransDetail := types.WeeklyTransDetail{
			Date:          showTime,
			NcTotalAmount: ncOrderAmount,
			ZfTotalAmount: zfOrderAmount,
			DfTotalAmount: dfOrderAmount,
			XfTotalAmount: xfOrderAmount,
		}
		weeklyTransDetails = append(weeklyTransDetails, weeklyTransDetail)

		//weeklyTransDetails, errAll := reportService.GetWeeklyOrderAmount(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		//if errAll != nil {
		//	return nil, errAll
		//}
		//
		//if len(weeklyTransDetails) > 0  {
		//	weeklyTransDetails[0].Date = showTime
		//	weeklyTransDetailsF = append(weeklyTransDetailsF, weeklyTransDetails[0])
		//}else {
		//	var weeklyTransDetail types.WeeklyTransDetail
		//	weeklyTransDetail.Date = showTime
		//	weeklyTransDetail.DfTotalAmount = 0.0
		//	weeklyTransDetail.XfTotalAmount = 0.0
		//	weeklyTransDetail.ZfTotalAmount = 0.0
		//	weeklyTransDetail.NcTotalAmount = 0.0
		//	weeklyTransDetailsF = append(weeklyTransDetailsF, weeklyTransDetail)
		//}
	}

	resp = &types.WeeklyTransDetailResponse{
		List: weeklyTransDetails,
	}

	return
}
