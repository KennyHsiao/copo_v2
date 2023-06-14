package bankblockaccount

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankBlockAccountCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankBlockAccountCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankBlockAccountCreateLogic {
	return BankBlockAccountCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankBlockAccountCreateLogic) BankBlockAccountCreate(req types.BankBlockAccountCreateRequest) error {
	//JWT取得登入账号
	account := l.ctx.Value("account").(string)
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		ux := model.NewBankBlockAccount(l.svcCtx.MyDB)
		isExist, err := ux.CheckIsBlockAccount(req.BankAccount)
		if err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		if isExist {
			return errorz.New(response.BANK_ACCOUNT_IN_BLACK_LIST)
		}
		bankBlockAccount := &types.BankBlockAccountCreate{
			BankBlockAccountCreateRequest: req,
			CreatedBy:                     account,
			UpdatedBy:                     account,
		}
		if err := l.svcCtx.MyDB.Table("bk_block_account").Create(bankBlockAccount).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}
