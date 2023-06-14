package merchantbindbank

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBindBankUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBindBankUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBindBankUpdateLogic {
	return MerchantBindBankUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBindBankUpdateLogic) MerchantBindBankUpdate(req types.MerchantBindBankUpdateRequest) error {
	merchantBindBankUpdate := types.MerchantBindBankUpdate{
		MerchantBindBankUpdateRequest: req,
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		if err := l.svcCtx.MyDB.Table("mc_merchant_bind_bank").Updates(merchantBindBankUpdate).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}
