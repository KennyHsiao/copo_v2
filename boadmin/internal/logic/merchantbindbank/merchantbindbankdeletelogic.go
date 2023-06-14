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

type MerchantBindBankDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBindBankDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBindBankDeleteLogic {
	return MerchantBindBankDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBindBankDeleteLogic) MerchantBindBankDelete(req types.MerchantBindBankDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		if err := l.svcCtx.MyDB.Table("mc_merchant_bind_bank").Delete(&req).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}
