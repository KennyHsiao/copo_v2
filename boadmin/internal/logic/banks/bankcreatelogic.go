package banks

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type BankCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankCreateLogic {
	return BankCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankCreateLogic) BankCreate(req types.BankCreateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		bank := &types.BankCreate{
			BankCreateRequest: req,
		}

		//检查是否有重复银行名
		isDuplicated, dupBank := model.CheckBankIsDuplicated(db, req)
		if isDuplicated {
			return errors.New(fmt.Sprintf(string("'%s:%s。 %+v'"), response.BANK_IS_DUPLICATED, "银行资料重复", dupBank))
		}
		//取得新银行代码
		if req.BankNo == "" {
			bankNo := model.GetNewBankCode(l.svcCtx.MyDB, req.CurrencyCode)
			length := 3 - len(bankNo)
			for i := 0; i < length; i++ {
				bankNo = "0" + bankNo
			}
			bank.BankNo = bankNo
		} else {
			bank.BankNo = req.BankNo
		}
		logx.Info("新增銀行資料:", bank)
		return db.Table("bk_banks").Create(bank).Error
	})
}
