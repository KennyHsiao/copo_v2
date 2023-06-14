package orderrecord

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

type ConfirmProxyPayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmProxyPayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) ConfirmProxyPayOrderLogic {
	return ConfirmProxyPayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConfirmProxyPayOrderLogic) ConfirmProxyPayOrder(req *types.ConfirmProxyPayOrderRequest) error {
	var rpcRequest transaction.ConfirmProxyPayOrderRequest
	copier.Copy(&rpcRequest, &req)
	// CALL transactionc
	rpcResp, err2 := l.svcCtx.TransactionRpc.ConfirmProxyPayOrderTransaction(l.ctx, &rpcRequest)
	if err2 != nil {
		return err2
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "ConfirmProxyPayOrderTransaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
