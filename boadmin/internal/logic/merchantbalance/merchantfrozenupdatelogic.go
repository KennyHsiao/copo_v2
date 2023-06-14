package merchantbalance

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/jinzhu/copier"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantFrozenUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantFrozenUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantFrozenUpdateLogic {
	return MerchantFrozenUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantFrozenUpdateLogic) MerchantFrozenUpdate(req types.MerchantFrozenUpdateRequest) error {
	IsAdmin := l.ctx.Value("isAdmin").(bool)

	if !IsAdmin {
		return errorz.New(response.ILLEGAL_REQUEST)
	}

	if isEnable, err := model.NewMerchantCurrency(l.svcCtx.MyDB).IsEnableDisplayPtBalance(req.MerchantCode, req.CurrencyCode); err != nil {
		return errorz.New(response.ILLEGAL_REQUEST)
	} else if isEnable && req.MerchantPtBalanceId == 0 {
		return errorz.New("请选择子钱包")
	} else if !isEnable && req.MerchantPtBalanceId != 0 {
		return errorz.New("禁止选择子钱包")
	}

	var rpcRequest transaction.MerchantBalanceFreezeRequest
	copier.Copy(&rpcRequest, &req)
	rpcRequest.UserAccount = l.ctx.Value("account").(string)
	// CALL transactionc
	rpcResp, err := l.svcCtx.TransactionRpc.MerchantBalanceFreezeTranaction(l.ctx, &rpcRequest)
	if err != nil {
		return err
	} else if rpcResp == nil {
		return errorz.New(response.SERVICE_RESPONSE_DATA_ERROR, "MerchantBalanceUpdateTranaction rpcResp is nil")
	} else if rpcResp.Code != response.API_SUCCESS {
		return errorz.New(rpcResp.Code, rpcResp.Message)
	}

	return nil
}
