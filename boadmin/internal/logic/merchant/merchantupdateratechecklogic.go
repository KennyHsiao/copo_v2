package merchant

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateRateCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateRateCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateRateCheckLogic {
	return MerchantUpdateRateCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateRateCheckLogic) MerchantUpdateRateCheck(req *types.MerchantUpdateRateCheckRequest) error {
	err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.Code).
		Update("rate_check", req.RateCheck).Error

	return err
}
