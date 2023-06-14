package merchantbindbank

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"
	"regexp"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBindBankCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBindBankCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBindBankCreateLogic {
	return MerchantBindBankCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBindBankCreateLogic) MerchantBindBankCreate(req types.MerchantBindBankCreateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {

		//验证银行卡号(必填)(必须为数字)(长度必须在10~22码)
		isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.BankAccount)
		currencyCode := req.CurrencyCode
		if currencyCode == constants.CURRENCY_THB {
			if req.BankAccount == "" || len(req.BankAccount) < 10 || len(req.BankAccount) > 16 || !isMatch2 {
				logx.Error("銀行卡號檢查錯誤，需10-16碼內：", req.BankAccount)
				return errorz.New(response.INVALID_BANK_NO, "BankAccount: "+req.BankAccount)
			}
		} else if currencyCode == constants.CURRENCY_CNY {
			if req.BankAccount == "" || len(req.BankAccount) < 13 || len(req.BankAccount) > 22 || !isMatch2 {
				logx.Error("銀行卡號檢查錯誤，需13-22碼內：", req.BankAccount)
				return errorz.New(response.INVALID_BANK_NO, "BankAccount: "+req.BankAccount)
			}
		}

		merchantBindBank := &types.MerchantBindBankCreate{
			MerchantBindBankCreateRequest: req,
		}
		// 从登入资讯渠得merchant_code
		merchantBindBank.MerchantCode = l.ctx.Value("merchantCode").(string)
		if err := l.svcCtx.MyDB.Table("mc_merchant_bind_bank").Create(merchantBindBank).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
	return nil
}
