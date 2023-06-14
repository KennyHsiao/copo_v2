package systemrate

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemRateUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemRateUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemRateUpdateLogic {
	return SystemRateUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemRateUpdateLogic) SystemRateUpdate(req *types.SystemRateUpdateRequest) error {
	userAccount := l.ctx.Value("account").(string)
	systemrate := types.SystemRate{
		ID:                  req.ID,
		WithdrawHandlingFee: req.WithdrawHandlingFee,
		MinWithdrawCharge:   req.MinWithdrawCharge,
		MaxWithdrawCharge:   req.MaxWithdrawCharge,
		CurrencyCode:        req.CurrencyCode,
		CreatedBy:           userAccount,
		UpdatedBy:           userAccount,
	}

	systemRateCreate := &types.SystemRateCeate{
		SystemRate: systemrate,
	}

	if req.ID > 0 {
		if err := l.svcCtx.MyDB.Table("bs_system_rate").
			Where("id = ?", systemrate.ID).
			Updates(map[string]interface{}{"withdraw_handling_fee": systemrate.WithdrawHandlingFee, "min_withdraw_charge": systemrate.MinWithdrawCharge, "max_withdraw_charge": systemrate.MaxWithdrawCharge}).Error; err != nil {
			return errorz.New(response.CREATE_FAILURE, err.Error())
		}
	} else {
		if err := l.svcCtx.MyDB.Table("bs_system_rate").
			Create(systemRateCreate).Error; err != nil {
			return errorz.New(response.CREATE_FAILURE, err.Error())
		}
	}

	return nil
}
