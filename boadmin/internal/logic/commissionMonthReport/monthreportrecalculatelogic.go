package commissionMonthReport

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MonthReportRecalculateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonthReportRecalculateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportRecalculateLogic {
	return MonthReportRecalculateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonthReportRecalculateLogic) MonthReportRecalculate(req *types.CommissionMonthReportRecalculateRequest) error {

	var rpcRequest transaction.RecalculateCommissionMonthReportRequest
	copier.Copy(&rpcRequest, &req)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.RecalculateCommissionMonthReport(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "RecalculateCommissionMonthReport rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
