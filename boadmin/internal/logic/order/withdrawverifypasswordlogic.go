package order

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawVerifyPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawVerifyPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawVerifyPasswordLogic {
	return WithdrawVerifyPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawVerifyPasswordLogic) WithdrawVerifyPassword(req types.WithdrawVerifyPasswordRequest) error {
	// JWT取得商户号资讯
	merchantCode := l.ctx.Value("merchantCode").(string)
	var merchant types.Merchant
	if err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND)
		}
		return errorz.New(response.DATABASE_FAILURE)
	}

	if len(merchant.WithdrawPassword) < 0 {
		return errorz.New(response.WITHDRAW_PASSWORD_NOT_SETTING)
	}

	if !utils.CheckPassword2(req.WithdrawPassword, merchant.WithdrawPassword) {
		return errorz.New(response.INCORRECT_WITHDRAW_PASSWD)
	}

	return nil
}
