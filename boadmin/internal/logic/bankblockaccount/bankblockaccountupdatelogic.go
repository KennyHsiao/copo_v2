package bankblockaccount

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankBlockAccountUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankBlockAccountUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankBlockAccountUpdateLogic {
	return BankBlockAccountUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankBlockAccountUpdateLogic) BankBlockAccountUpdate(req types.BankBlockAccountUpdateRequest) error {
	//JWT取得登入账号
	account := l.ctx.Value("account").(string)
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		bankBlockAccount := &types.BankBlockAccountUpdate{
			BankBlockAccountUpdateRequest: req,
			UpdatedBy:                     account,
		}
		if err = l.svcCtx.MyDB.Table("bk_block_account").Updates(bankBlockAccount).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})
}
