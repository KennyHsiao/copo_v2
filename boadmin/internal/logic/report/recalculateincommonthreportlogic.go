package report

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReCalculateIncomMonthReportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReCalculateIncomMonthReportLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReCalculateIncomMonthReportLogic {
	return ReCalculateIncomMonthReportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReCalculateIncomMonthReportLogic) ReCalculateIncomMonthReport(req *types.ReCalculateIncomMonthReportRequest) error {

	rpcRequest := transaction.CalculateMonthProfitReportRequest{
		Month: req.Month,
	}

	rpcResp, err := l.svcCtx.TransactionRpc.CalculateMonthProfitReport(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "RecalculateIncomMonthReport rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}
	return nil
}
