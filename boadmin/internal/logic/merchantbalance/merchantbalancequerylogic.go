package merchantbalance

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceQueryLogic {
	return MerchantBalanceQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceQueryLogic) MerchantBalanceQuery(req types.MerchantBalanceQueryRequest) (resp *types.MerchantBalanceQueryResponse, err error) {

	err = l.svcCtx.MyDB.Table("mc_merchant_balances").Take(&resp, req.ID).Error

	return
}
