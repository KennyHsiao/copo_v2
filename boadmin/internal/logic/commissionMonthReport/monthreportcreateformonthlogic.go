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

type MonthReportCreateForMonthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonthReportCreateForMonthLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportCreateForMonthLogic {
	return MonthReportCreateForMonthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonthReportCreateForMonthLogic) MonthReportCreateForMonth(req *types.CommissionMonthReportCreateForMonthRequest) error {
	var rpcRequest transaction.CalculateCommissionMonthAllRequest
	copier.Copy(&rpcRequest, &req)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.CalculateCommissionMonthAllReport(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "RecalculateCommissionMonthReport rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
