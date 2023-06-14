package banks

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankDeleteLogic {
	return BankDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankDeleteLogic) BankDelete(req types.BankDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		if err := db.Table("bk_banks").Delete(&req).Error; err != nil {
			logx.Error("err:", err)
			return errorz.New(response.DELETE_DATABASE_FAILURE, err.Error())
		}
		return err
	})
}
