package merchantbalance

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceUpdateLogic {
	return MerchantBalanceUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceUpdateLogic) MerchantBalanceUpdate(req types.MerchantBalanceUpdateRequest) error {

	IsAdmin := l.ctx.Value("isAdmin").(bool)

	if !IsAdmin {
		return errorz.New(response.ILLEGAL_REQUEST)
	}

	if isEnable, err := model.NewMerchantCurrency(l.svcCtx.MyDB).IsEnableDisplayPtBalance(req.MerchantCode, req.CurrencyCode); err != nil {
		return errorz.New(response.ILLEGAL_REQUEST)
	} else if isEnable {
		return errorz.New(response.SUB_WALLET_ENABLED_THEREFORE_OPERATION_PROHIBITED)
	}

	var rpcRequest transaction.MerchantBalanceUpdateRequest
	copier.Copy(&rpcRequest, &req)
	rpcRequest.UserAccount = l.ctx.Value("account").(string)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.MerchantBalanceUpdateTranaction(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "MerchantBalanceUpdateTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
