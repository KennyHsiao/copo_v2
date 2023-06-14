package banks

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankUpdateLogic {
	return BankUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankUpdateLogic) BankUpdate(req types.BankUpdateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		//检查银行是否已经存在
		err = model.CheckBankIsExist(db, req)
		if err != nil {
			return errors.New(err.Error())
		}
		//检查是否有重复银行名
		isDuplicated := model.CheckBankIsDuplicatedUp(db, req)
		if isDuplicated {
			return errors.New(response.BANK_IS_DUPLICATED)
		}

		updateBank := &types.BankUpdate{
			BankUpdateRequest: req,
		}

		return db.Table("bk_banks").Where("bank_no = ? AND currency_code = ?", req.BankNo, req.CurrencyCode).Updates(updateBank).Error
	})
}
