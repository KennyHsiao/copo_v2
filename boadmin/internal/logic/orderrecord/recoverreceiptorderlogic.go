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

type RecoverReceiptOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecoverReceiptOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) RecoverReceiptOrderLogic {
	return RecoverReceiptOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecoverReceiptOrderLogic) RecoverReceiptOrder(req *types.RecoverReceiptOrderRequest) error {
	var rpcRequest transaction.RecoverReceiptOrderRequest
	copier.Copy(&rpcRequest, &req)
	rpcRequest.UserAccount = l.ctx.Value("account").(string)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.RecoverReceiptOrderTransaction(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "RecoverReceiptOrderTransaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
