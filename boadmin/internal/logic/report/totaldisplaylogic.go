package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/common/utils"
	"context"
	"sync"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TotalDisplayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTotalDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) TotalDisplayLogic {
	return TotalDisplayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TotalDisplayLogic) TotalDisplay(req *types.TotalDisplayRequest) (resp *types.TotalDisplayResponse, err error) {
	db := l.svcCtx.MyDB

	nowTime := time.Now()
	startTime, endTime := reportService.GetQueryTodayTime(nowTime)
	var totalMerchantBalances *types.TotalMerchantBalances
	var ncTotalAmount float64
	var zfTotalAmount float64
	var dfTotalAmount float64
	var xfTotalAmount float64
	var incomeDetail types.IncomeDetailResponse
	var err1, err2, err3, err4, err5, err6 error
	var ncRate, zfRate, dfRate, xfHandlingFee float64

	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		// 商户目前，代付、下发总馀额
		totalMerchantBalances, err1 = reportService.GetTotalMerchantBalance(db, req.CurrencyCode, l.ctx)
	}()

	go func() {
		defer wg.Done()
		// 取得总订单金额
		ncTotalAmount, err2 = reportService.GetOneOrderTypeTotalAmount(db, "NC", req.CurrencyCode, startTime, endTime, l.ctx)
	}()

	go func() {
		defer wg.Done()
		zfTotalAmount, err3 = reportService.GetOneOrderTypeTotalAmount(db, "ZF", req.CurrencyCode, startTime, endTime, l.ctx)
	}()

	go func() {
		defer wg.Done()
		dfTotalAmount, err4 = reportService.GetOneOrderTypeTotalAmount(db, "DF", req.CurrencyCode, startTime, endTime, l.ctx)

	}()
	go func() {
		defer wg.Done()
		xfTotalAmount, err5 = reportService.GetOneOrderTypeTotalAmount(db, "XF", req.CurrencyCode, startTime, endTime, l.ctx)
	}()
	// 取得系统利润
	go func() {
		defer wg.Done()
		ncRate, err6 = reportService.GetPerDayIncomeDetailNc(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		zfRate, err6 = reportService.GetPerDayIncomeDetailZf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		dfRate, err6 = reportService.GetPerDayIncomeDetailDf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		xfHandlingFee, err6 = reportService.GetPerDayIncomeDetailXf(l.svcCtx.MyDB, req.CurrencyCode, startTime, endTime, l.ctx)
		incomeDetail.NcRate = ncRate
		incomeDetail.ZfRate = zfRate
		incomeDetail.DfRate = dfRate
		incomeDetail.XfHandlingFee = xfHandlingFee
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
	if err5 != nil {
		return nil, err5
	}
	if err6 != nil {
		return nil, err6
	}
	receiveTransAmount := utils.FloatAdd(ncTotalAmount, zfTotalAmount)
	payoutTransAmount := utils.FloatAdd(dfTotalAmount, xfTotalAmount)

	totalSystemAmount := utils.FloatAdd(incomeDetail.NcRate, utils.FloatAdd(incomeDetail.ZfRate, utils.FloatAdd(incomeDetail.DfRate, incomeDetail.XfHandlingFee)))

	resp = &types.TotalDisplayResponse{
		ProxyPayBalance:    totalMerchantBalances.DfBalances,
		WithdrawBalance:    totalMerchantBalances.XfBalances,
		ReceiveTransAmount: receiveTransAmount,
		PayoutTransAmount:  payoutTransAmount,
		Income:             totalSystemAmount,
	}
	return
}
