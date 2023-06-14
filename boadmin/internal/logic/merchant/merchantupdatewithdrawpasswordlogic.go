package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateWithdrawPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateWithdrawPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateWithdrawPasswordLogic {
	return MerchantUpdateWithdrawPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateWithdrawPasswordLogic) MerchantUpdateWithdrawPassword(req types.UpdateWithdrawPasswordRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		var merchant types.Merchant
		merchantCode := l.ctx.Value("merchantCode").(string)

		if len(merchantCode) == 0 {
			return errorz.New(response.SETTING_FAILURE)
		}

		if err = db.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		merchant.WithdrawPassword = utils.PasswordHash2(req.WithdrawPassword)
		merchant.IsWithdraw = "1"

		if err = db.Table("mc_merchants").Updates(types.MerchantX{
			Merchant: merchant,
		}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})

}
