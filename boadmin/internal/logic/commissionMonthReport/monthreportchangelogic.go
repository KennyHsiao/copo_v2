package commissionMonthReport

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MonthReportChangeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonthReportChangeLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportChangeLogic {
	return MonthReportChangeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonthReportChangeLogic) MonthReportChange(req *types.CommissionMonthReportChangeRequest) error {
	var report types.CommissionMonthReportX
	// 取得報表
	if err := l.svcCtx.MyDB.Table("cm_commission_month_reports").Where("id = ?", req.ID).Find(&report).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if report.Status == "1" {
		// 已審核報表不可再異動
		return errorz.New(response.MERCHANT_COMMISSION_AUDIT)
	}

	report.ChangeCommission = req.ChangeCommission
	report.Comment = req.Comment
	if err := l.svcCtx.MyDB.Table("cm_commission_month_reports").Updates(&report).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
