package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantResetWithdrawPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantResetWithdrawPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantResetWithdrawPasswordLogic {
	return MerchantResetWithdrawPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantResetWithdrawPasswordLogic) MerchantResetWithdrawPassword(req types.ResetWithdrawPasswordRequest) (err error) {
	merchant := types.Merchant{
		ID:               req.ID,
		WithdrawPassword: "0",
		IsWithdraw:       "0",
	}

	if err = l.svcCtx.MyDB.Table("mc_merchants").Updates(types.MerchantX{
		Merchant: merchant,
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return nil
}
