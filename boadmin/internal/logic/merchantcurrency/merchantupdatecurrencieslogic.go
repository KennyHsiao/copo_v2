package merchantcurrency

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type MerchantUpdateCurrenciesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateCurrenciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateCurrenciesLogic {
	return MerchantUpdateCurrenciesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateCurrenciesLogic) MerchantUpdateCurrencies(req types.MerchantUpdateCurrenciesRequestX) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		var merchant *types.Merchant
		var merchants []types.Merchant
		var enableCurrencies []string

		// 取得當前商戶
		if merchant, err = model.NewMerchant(db).GetMerchantByCode(req.MerchantCode); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 取得子商戶(包含自己)
		if merchants, err = getSubagentMerchants(db, *merchant); err != nil {
			return errorz.New(response.DATABASE_FAILURE, "取得子商戶: "+err.Error())
		}

		for _, currency := range req.Currencies {

			if currency.Status == "1" {
				currency.MerchantCode = req.MerchantCode
				updateCurrency(db, currency)
				enableCurrencies = append(enableCurrencies, currency.CurrencyCode)
			} else {
				// 禁用幣別需考慮子代理
				for _, mrct := range merchants {
					merchantCurrency := currency
					merchantCurrency.MerchantCode = mrct.Code
					updateCurrency(db, merchantCurrency)
				}
			}
		}
		// 編輯每個商戶下的每個帳戶可見幣別
		updateUsersCurrencies(db, enableCurrencies, merchants)

		return err
	})
}

func updateCurrency(db *gorm.DB, currency types.MerchantCurrencyUpdate) (err error) {

	var originalCurrency *types.MerchantCurrency
	//查詢是否新增過 並給予ID
	if err = db.Table("mc_merchant_currencies").
		Where("merchant_code = ? AND currency_code = ?", currency.MerchantCode, currency.CurrencyCode).
		Find(&originalCurrency).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	currency.ID = originalCurrency.ID

	mcx := model.NewMerchantCurrency(db)
	mbx := model.NewMerchantBalance(db)

	//if currency.CurrencyCode == "CNY" {
	//	// 人民幣不能禁用 但可以編輯下發手續費
	//	currency.Status = "1"
	//}

	if currency.ID != 0 {
		// 編輯幣別
		if err = db.Table("mc_merchant_currencies").Select("id", "merchant_code", "currency_code", "withdraw_handling_fee", "status", "sort_order", "update_at").
			Updates(currency).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	} else {
		// 新增幣別
		if err = mcx.CreateMerchantCurrency(currency.MerchantCode, currency.CurrencyCode, currency.Status, currency.SortOrder); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		// 新增幣別錢包
		if err = mbx.CreateMerchantBalances(currency.MerchantCode, currency.CurrencyCode); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	return
}

func getSubagentMerchants(db *gorm.DB, merchant types.Merchant) ([]types.Merchant, error) {
	var merchants []types.Merchant
	var err error

	// 是否為代理商
	if len(merchant.AgentLayerCode) > 0 {
		// 取得自己和子孫商戶
		ux := model.NewMerchant(db)
		if merchants, err = ux.GetDescendantAgents(merchant.AgentLayerCode, true); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	} else {
		// 不是代理只回傳自己
		merchants = append(merchants, merchant)
	}
	return merchants, err
}

func updateUsersCurrencies(db *gorm.DB, currencies []string, merchants []types.Merchant) (err error) {
	// 每個子商戶(包含自己)
	for _, merchant := range merchants {
		// 商戶下每個帳號
		for _, user := range merchant.Users {
			sameCurrencies := getSameCurrencies(user.Currencies, currencies)
			db.Model(&user).Association("Currencies").Clear()
			user.Currencies = sameCurrencies

			if err = db.Table("au_users").Updates(user).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}
	}
	return
}

func getSameCurrencies(originalCurrencies []types.Currency, currencies []string) []types.Currency {
	var finalarray []types.Currency
	for _, a1 := range originalCurrencies {
		for _, a2 := range currencies {
			if a1.Code == a2 {
				finalarray = append(finalarray, a1)
			}
		}
	}
	return finalarray
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
