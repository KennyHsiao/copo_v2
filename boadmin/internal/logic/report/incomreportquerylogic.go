package report

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncomReportQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomReportQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomReportQueryLogic {
	return IncomReportQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomReportQueryLogic) IncomReportQuery(req *types.IncomReportMonthQueryRequest) (resp *types.IcomReportMonthQueryResponse, err error) {
	var incomReport []types.IncomReport

	if err = l.svcCtx.MyDB.Table("rp_incom_report").
		Where("month >= ? AND month <= ?", req.StartMonth, req.EndMonth).
		Where("currency_code = ?", req.CurrencyCode).
		Find(&incomReport).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.IcomReportMonthQueryResponse{
		List: incomReport,
	}

	return resp, nil
}
