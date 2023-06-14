package order

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawVerifyWayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawVerifyWayLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawVerifyWayLogic {
	return WithdrawVerifyWayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawVerifyWayLogic) WithdrawVerifyWay() (resp *types.WithdrawVerifyWayResponse, err error) {
	//JWT取得商户资讯
	merchantCode := l.ctx.Value("merchantCode").(string)
	var merchant types.Merchant

	if err = l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.DATA_NOT_FOUND)
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}
	}

	if merchant.PayingValidatedType == constants.PAYING_VALIDATED_TYPE_PASSWORD &&
		merchant.WithdrawPassword == "" {
		return nil, errorz.New(response.WITHDRAW_PASSWORD_NOT_SETTING)
	}
	resp = &types.WithdrawVerifyWayResponse{
		PayingValidatedType: merchant.PayingValidatedType,
	}
	return
}
