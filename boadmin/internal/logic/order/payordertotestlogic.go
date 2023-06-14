package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/common/constants"
	"github.com/copo888/transaction_service/rpc/transactionclient"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayOrderToTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayOrderToTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayOrderToTestLogic {
	return PayOrderToTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayOrderToTestLogic) PayOrderToTest(req *types.PayOrderToTestRequest) (resp *types.PayOrderToTestResponse, err error) {

	txOrder := &types.OrderX{}
	if txOrder, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.OrderNo, ""); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有成功單可以轉測試單
	if txOrder.Status != constants.SUCCESS {
		return nil, errorz.New(response.ORDER_STATUS_WRONG, "")
	}
	// 鎖訂單不可轉正式單
	if txOrder.IsLock == constants.IS_LOCK_YES {
		return nil, errorz.New(response.ORDER_IS_STATUS_IS_LOCK, "")
	}
	// 補單,追回單...特殊單不可轉測試單
	if len(txOrder.ReasonType) > 0 {
		return nil, errorz.New(response.ORDER_IS_STATUS_IS_LOCK, "")
	}

	rpcResp, errRpc := l.svcCtx.TransactionRpc.PayOrderSwitchTest(l.ctx, &transactionclient.PayOrderSwitchTestRequest{
		OrderNo:     txOrder.OrderNo,
		UserAccount: l.ctx.Value("account").(string),
	})
	if errRpc != nil {
		return nil, errorz.New(response.SYSTEM_ERROR, errRpc.Error())
	} else if rpcResp == nil {
		return nil, errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return nil, errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return
}
