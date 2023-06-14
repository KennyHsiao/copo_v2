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

type BankBlockAccountDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankBlockAccountDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankBlockAccountDeleteLogic {
	return BankBlockAccountDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankBlockAccountDeleteLogic) BankBlockAccountDelete(req types.BankBlockAccountDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		if err = l.svcCtx.MyDB.Table("bk_block_account").Delete(&req).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})
}
