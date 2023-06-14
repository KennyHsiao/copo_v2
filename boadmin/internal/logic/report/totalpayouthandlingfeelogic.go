package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
)

type TotalPayoutHandlingFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTotalPayoutHandlingFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) TotalPayoutHandlingFeeLogic {
	return TotalPayoutHandlingFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TotalPayoutHandlingFeeLogic) TotalPayoutHandlingFee(req *types.TotalPayoutHandlingFeeRequest) (resp *types.TotalPayoutHandlingFeeResponse, err error) {
	var totalPayouts []types.TotalPayout
	//var cstZone, loadErr = time.LoadLocation(req.Location)
	//if loadErr != nil {
	//	return nil, errorz.New(response.INVALID_TIMESTAMP, loadErr.Error())
	//}
	nowTime := time.Now()
	//st := nowTime.AddDate(0, 0, -7).Format("2006-01-02")
	//startTimeStr := fmt.Sprint(st) + " 16:00:00"
	//et := nowTime.Format("2006-01-02")
	//endTimeStr := fmt.Sprintf(et) + " 16:00:00"
	//totalPayouts, err = reportService.GetPerDayIncomeDetail2(l.svcCtx.MyDB, req.CurrencyCode, startTimeStr, endTimeStr, l.ctx)
	//if err != nil {
	//	return nil, err
	//}
	//
	//var showTimes []string
	//for i := 0; i < 7; i++ {
	//	_, _, showTime := reportService.GetWeeklyDate(i, nowTime)
	//	showTimes = append(showTimes, showTime)
	//}
	//var payoutMap map[string]types.TotalPayout
	//payoutMap = make(map[string]types.TotalPayout)
	//var totalPayoutsF []types.TotalPayout
	//if len(totalPayouts) > 0 {
	//	for i, _ := range totalPayouts {
	//		ds := strings.Split(totalPayouts[i].Date, "T")
	//		totalPayouts[i].Date = ds[0]
	//		payoutMap[totalPayouts[i].Date]=totalPayouts[i]
	//	}
	//	for _, showTime := range showTimes {
	//		v, ise := payoutMap[showTime]
	//		if ise != true {
	//			tp := types.TotalPayout{
	//				Date: showTime,
	//				ProxyPayHandlingFee: 0.0,
	//				WithdrawHandlingFee: 0.0,
	//			}
	//			totalPayoutsF = append(totalPayoutsF, tp)
	//		}else {
	//			totalPayoutsF = append(totalPayoutsF, v)
	//		}
	//	}
	//}else{
	//	for _, showTime := range showTimes {
	//		tp := types.TotalPayout{
	//			Date: showTime,
	//			ProxyPayHandlingFee: 0.0,
	//			WithdrawHandlingFee: 0.0,
	//		}
	//		totalPayoutsF = append(totalPayoutsF, tp)
	//	}
	//}

	for i := 0; i < 7; i++ {
		startTime, endTime, showTime := reportService.GetWeeklyDate(i, nowTime)
		//var incomeDetail *types.IncomeDetailResponse
		//incomeDetail, err = reportService.GetPerDayIncomeDetail(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime)
		var dfRate, xfHandlingFee float64
		var errDf, errXf error
		dfRate, errDf = reportService.GetPerDayIncomeDetailDf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		if errDf != nil {
			return nil, errDf
		}
		xfHandlingFee, errXf = reportService.GetPerDayIncomeDetailXf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		if errXf != nil {
			return nil, errXf
		}

		totalPayout := types.TotalPayout{
			Date:                showTime,
			ProxyPayHandlingFee: dfRate,
			WithdrawHandlingFee: xfHandlingFee,
		}
		totalPayouts = append(totalPayouts, totalPayout)
	}

	resp = &types.TotalPayoutHandlingFeeResponse{
		List: totalPayouts,
	}

	return
}
