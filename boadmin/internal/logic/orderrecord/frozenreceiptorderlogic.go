package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type FrozenReceiptOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFrozenReceiptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) FrozenReceiptOrderLogic {
	return FrozenReceiptOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FrozenReceiptOrderLogic) FrozenReceiptOrder(req types.FrozenReceiptOrderRequest) error {
	var rpcRequest transaction.FrozenReceiptOrderRequest
	copier.Copy(&rpcRequest, &req)
	rpcRequest.UserAccount = l.ctx.Value("account").(string)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.FrozenReceiptOrderTransaction(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "FrozenReceiptOrderTransaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
