package merchantcurrency

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantGetAvailableCurrencyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantGetAvailableCurrencyLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantGetAvailableCurrencyLogic {
	return MerchantGetAvailableCurrencyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantGetAvailableCurrencyLogic) MerchantGetAvailableCurrency(req types.MerchantGetAvailableCurrencyRequest) (merchantCurrencies []types.MerchantCurrency, err error) {
	var merchant *types.Merchant

	// 取得商戶資料
	if merchant, err = model.NewMerchant(l.svcCtx.MyDB).GetMerchantByCode(req.MerchantCode); errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorz.New(response.INVALID_MERCHANT_CODING, err.Error())
	} else if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	// 取得商戶幣別
	if merchantCurrencies, err = model.NewMerchantCurrency(l.svcCtx.MyDB).GetByMerchantCode(merchant.Code, ""); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if merchant.AgentParentCode == "" {
		merchantCurrencies, err = GetAvailableCurrencyByGeneralAgent(l.svcCtx.MyDB, merchantCurrencies, merchant)
	} else {
		merchantCurrencies, err = GetAvailableCurrencyBySubagent(l.svcCtx.MyDB, merchantCurrencies, merchant)
	}

	return merchantCurrencies, err
}

// GetAvailableCurrencyByGeneralAgent 無上層代理可用幣別為全部
func GetAvailableCurrencyByGeneralAgent(db *gorm.DB, merchantCurrencies []types.MerchantCurrency, merchant *types.Merchant) ([]types.MerchantCurrency, error) {
	var currencies []types.Currency
	var availableMerchantCurrencies []types.MerchantCurrency
	// 可用幣別為全部幣別
	if err := db.Table("bs_currencies").Order("code").Find(&currencies).Error; err != nil {
		return nil, err
	}

	// 把礎資料幣別轉換成乾淨幣別選項
	for _, currency := range currencies {
		availableMerchantCurrencies = append(availableMerchantCurrencies, types.MerchantCurrency{
			MerchantCode: merchant.Code,
			CurrencyCode: currency.Code,
			Status:       "0",
		})
	}

	// 若商戶已有此幣別 則用商戶幣別替換選項
	for i, availableMerchantCurrency := range availableMerchantCurrencies {
		for _, merchantCurrency := range merchantCurrencies {
			if availableMerchantCurrency.CurrencyCode == merchantCurrency.CurrencyCode {
				availableMerchantCurrencies[i] = merchantCurrency
				break
			}
		}
	}
	return availableMerchantCurrencies, nil
}

// GetAvailableCurrencyBySubagent 有上層代理 可用幣別為上層已啟用幣別
func GetAvailableCurrencyBySubagent(db *gorm.DB, merchantCurrencies []types.MerchantCurrency, merchant *types.Merchant) ([]types.MerchantCurrency, error) {

	var availableMerchantCurrencies []types.MerchantCurrency

	// 可用幣別為上層已啟用幣別
	parentMerchantCurrencies, err := model.NewMerchantCurrency(db).GetByMerchantCode(merchant.AgentParentCode, "1")
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 將上層代理幣別 替換成乾淨幣別選項
	for _, parentMerchantCurrency := range parentMerchantCurrencies {
		availableMerchantCurrencies = append(availableMerchantCurrencies, types.MerchantCurrency{
			MerchantCode: merchant.Code,
			CurrencyCode: parentMerchantCurrency.CurrencyCode,
			Status:       "0",
		})
	}

	// 若商戶已有此幣別 則用商戶幣別替換選項
	for i, availableMerchantCurrency := range availableMerchantCurrencies {
		for _, merchantCurrency := range merchantCurrencies {
			if availableMerchantCurrency.CurrencyCode == merchantCurrency.CurrencyCode {
				availableMerchantCurrencies[i] = merchantCurrency
				break
			}
		}
	}
	return availableMerchantCurrencies, nil
}
