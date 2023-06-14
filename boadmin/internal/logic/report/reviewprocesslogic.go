package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type ReviewProcessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReviewProcessLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReviewProcessLogic {
	return ReviewProcessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReviewProcessLogic) ReviewProcess(req *types.ReviewProcessRequest) (resp *types.ReviewProcessResponse, err error) {
	resp = &types.ReviewProcessResponse{}

	db := l.svcCtx.MyDB

	nowTime := time.Now()
	startTime, endTime := reportService.GetQueryTodayTime(nowTime)

	var internalChargeOrderNum, withdrawOrderNum, personalRepaymentNum int
	var errNc, errXf, errPerson error
	internalChargeOrderNum, errNc = l.reviewProcessNc(db, startTime, endTime, req.CurrencyCode, l.ctx)
	if errNc != nil {
		return nil, errNc
	}
	withdrawOrderNum, errXf = l.reviewProcessXf(db, startTime, endTime, req.CurrencyCode, l.ctx)
	if errXf != nil {
		return nil, errXf
	}
	personalRepaymentNum, errPerson = l.reviewProcessPersonProcess(db, startTime, endTime, req.CurrencyCode, l.ctx)
	if errPerson != nil {
		return nil, errPerson
	}
	//db = db.Where("currency_code = ?", req.CurrencyCode)
	//db = db.Where("created_at >= ?", startTime)
	//db = db.Where("created_at < ?", endTime)
	//db = db.Where("type = 'NC'")
	//
	//selectX := "COUNT(CASE WHEN type = 'NC' THEN 1 END) AS internal_charge_order_num," +
	//	"COUNT(CASE WHEN type = 'XF' THEN 1 END) AS withdraw_order_num," +
	//	"COUNT(CASE WHEN person_process_status != '10' THEN 1 END) AS personal_repayment_num"
	//
	//if err = db.WithContext(l.ctx).Table("tx_orders").
	//	Select(selectX).Find(resp).Error; err != nil {
	//		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	resp = &types.ReviewProcessResponse{
		InternalChargeOrderNum: strconv.Itoa(internalChargeOrderNum),
		WithdrawOrderNum:       strconv.Itoa(withdrawOrderNum),
		PersonalRepaymentNum:   strconv.Itoa(personalRepaymentNum),
	}

	return
}

func (l *ReviewProcessLogic) reviewProcessNc(db *gorm.DB, stTime, edTime string, currencyCode string, ctx context.Context) (resp int, err error) {
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("created_at >= ?", stTime)
	db = db.Where("created_at < ?", edTime)
	db = db.Where("type = 'NC'")

	selectX := "COUNT(*) AS internal_charge_order_num"

	if err = db.WithContext(ctx).Table("tx_orders").
		Select(selectX).Find(&resp).Error; err != nil {
		return 0.0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l *ReviewProcessLogic) reviewProcessXf(db *gorm.DB, stTime, edTime string, currencyCode string, ctx context.Context) (resp int, err error) {
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("created_at >= ?", stTime)
	db = db.Where("created_at < ?", edTime)
	db = db.Where("type = 'XF'")

	selectX := "COUNT(*) AS internal_charge_order_num"

	if err = db.WithContext(ctx).Table("tx_orders").
		Select(selectX).Find(&resp).Error; err != nil {
		return 0.0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l *ReviewProcessLogic) reviewProcessPersonProcess(db *gorm.DB, stTime, edTime string, currencyCode string, ctx context.Context) (resp int, err error) {
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("created_at >= ?", stTime)
	db = db.Where("created_at < ?", edTime)
	db = db.Where("type = 'DF'")
	db = db.Where("person_process_status != '10'")

	selectX := "COUNT(*) AS internal_charge_order_num"

	if err = db.WithContext(ctx).Table("tx_orders").
		Select(selectX).Find(&resp).Error; err != nil {
		return 0.0, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
