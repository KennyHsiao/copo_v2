package systemrate

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemRateCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemRateCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemRateCreateLogic {
	return SystemRateCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemRateCreateLogic) SystemRateCreate(req *types.SystemRateRequest) error {
	userAccount := l.ctx.Value("account").(string)
	systemrate := types.SystemRate{
		CurrencyCode:        req.CurrencyCode,
		WithdrawHandlingFee: req.WithdrawHandlingFee,
		MinWithdrawCharge:   req.MinWithdrawCharge,
		MaxWithdrawCharge:   req.MaxWithdrawCharge,
		CreatedBy:           userAccount,
		UpdatedBy:           userAccount,
	}

	systemRateCreate := &types.SystemRateCeate{
		SystemRate: systemrate,
	}

	if err := l.svcCtx.MyDB.Table("bs_system_rate").Create(systemRateCreate).Error; err != nil {
		return errorz.New(response.CREATE_FAILURE, err.Error())
	}

	return nil
}
