package merchantbalance

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceCreateLogic {
	return MerchantBalanceCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceCreateLogic) MerchantBalanceCreate(req types.MerchantBalanceCreateRequest) error {
	merchantBalance := &types.MerchantBalanceCreate{
		MerchantBalanceCreateRequest: req,
	}
	return l.svcCtx.MyDB.Table("mc_merchant_balances").Create(merchantBalance).Error
}
