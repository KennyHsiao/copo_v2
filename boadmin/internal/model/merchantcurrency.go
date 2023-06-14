package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
)

type merchantCurrency struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchantCurrency(mydb *gorm.DB, t ...string) *merchantCurrency {
	table := "mc_merchant_currencies"
	if len(t) > 0 {
		table = t[0]
	}
	return &merchantCurrency{
		MyDB:  mydb,
		Table: table,
	}
}

func (c *merchantCurrency) GetByMerchantCode(code string, status string) (merchantCurrencies []types.MerchantCurrency, err error) {
	//var terms []string
	db := c.MyDB
	if len(status) > 0 {
		//terms = append(terms, fmt.Sprintf("status = '%s'", status))
		db = db.Where("status = ?", status)
	}
	//terms = append(terms, fmt.Sprintf("merchant_code  = '%s'", code))
	db = db.Where("merchant_code = ?", code)
	//term := strings.Join(terms, " AND ")

	err = db.Table(c.Table).
		Order("currency_code").
		Find(&merchantCurrencies).Error

	return
}

func (c *merchantCurrency) CreateMerchantCurrency(merchantCode, currencyCode, status string, sortOrder int64) (err error) {
	//var isExist bool
	//
	//if isExist, err = c.IsExistFromMerchantCurrency(merchantCode, currencyCode); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, err.Error())
	//} else if isExist {
	//	return
	//}

	if err = c.MyDB.Table(c.Table).Create(&types.MerchantCurrencyCreate{
		MerchantCurrencyCreateRequest: types.MerchantCurrencyCreateRequest{
			MerchantCode: merchantCode,
			CurrencyCode: currencyCode,
			SortOrder:    sortOrder,
			Status:       status,
		},
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (c *merchantCurrency) IsExistFromMerchantCurrency(merchantCode, currencyCode string) (isExist bool, err error) {
	err = c.MyDB.Table(c.Table).
		Select("count(*) > 0").
		Where("merchant_code = ? AND currency_code = ?", merchantCode, currencyCode).
		Find(&isExist).Error
	return
}

func (c *merchantCurrency) IsEnableDisplayPtBalance(merchantCode, currencyCode string) (isEnable bool, err error) {
	err = c.MyDB.Table(c.Table).
		Select("count(*) > 0").
		Where("merchant_code = ? AND currency_code = ? AND is_display_pt_balance = ?", merchantCode, currencyCode, "1").
		Find(&isEnable).Error
	return
}
