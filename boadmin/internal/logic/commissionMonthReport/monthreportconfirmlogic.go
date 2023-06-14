package commissionMonthReport

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type MonthReportConfirmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonthReportConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportConfirmLogic {
	return MonthReportConfirmLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonthReportConfirmLogic) MonthReportConfirm(req *types.CommissionMonthReportConfirmRequest) error {

	var rpcRequest transaction.ConfirmCommissionMonthReportRequest
	rpcRequest.ConfirmBy = l.ctx.Value("account").(string)
	rpcRequest.ID = req.ID
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.ConfirmCommissionMonthReport(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "ConfirmCommissionMonthReport rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}
	// CALL 重新計算收益報表
	location, _ := time.LoadLocation("Asia/Taipei")
	month := time.Now().In(location).Format("2006-01")

	rpcMonthProfitReq := transaction.CalculateMonthProfitReportRequest{
		Month: month,
	}

	rpcMonthProfitResp, err := l.svcCtx.TransactionRpc.CalculateMonthProfitReport(l.ctx, &rpcMonthProfitReq)
	if err != nil {
		return err
	} else if rpcMonthProfitResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "RecalculateMonthIncomReport rpcResp is nil")
	} else if rpcMonthProfitResp.Code != response.API_SUCCESS {
		return errorz.New(rpcMonthProfitResp.Code, rpcMonthProfitResp.Message)
	}

	return nil
}
