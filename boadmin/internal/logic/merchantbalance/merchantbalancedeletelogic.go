package merchantbalance

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceDeleteLogic {
	return MerchantBalanceDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceDeleteLogic) MerchantBalanceDelete(req types.MerchantBalanceDeleteRequest) error {
	return l.svcCtx.MyDB.Table("mc_merchant_balances").Delete(&req).Error
}
