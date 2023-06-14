package merchantPtBalance

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceDisableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceDisableLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceDisableLogic {
	return MerchantPtBalanceDisableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceDisableLogic) MerchantPtBalanceDisable(req *types.MerchantPtDisableRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
