package ordersService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"errors"
	"gorm.io/gorm"
)

func GetMerchantChannelRate(db *gorm.DB, merchantCode string, currencyCode string, orderType string) (resp []*types.MerchantOrderRateListViewX, err error) {
	//var merchantOrderRateListViews []*types.MerchantOrderRateListViewX
	//var terms []string
	//terms = append(terms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
	//terms = append(terms, "merchnrate_status = '1'")           // 0:禁用 1:啟用
	//terms = append(terms, fmt.Sprintf("pay_type_code = 'DF'")) // 內充看代付
	//terms = append(terms, "designation = '1'")
	//terms = append(terms, "chn_status = '1'")
	//terms = append(terms, "chnpaytype_status = '1'")
	//terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	db = db.Where("merchant_code = ?", merchantCode)
	db = db.Where("merchnrate_status = '1'")
	db = db.Where("designation = '1'")
	db = db.Where("chn_status = '1'")
	db = db.Where("chnpaytype_status = '1'")
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("pay_type_code = 'DF'")
	if orderType == "NC" {
		//terms = append(terms, fmt.Sprintf("chn_is_proxy = '0'")) //支援支轉代(0:不支援 1:支援)
		db = db.Where("chn_is_proxy = '0'")
	}

	//term := strings.Join(terms, " AND ")

	// 查询商户拥有的代付渠道
	// TODO 需要判斷商戶可用幣別的代付餘額下限值??(V1在系统常量)
	if orderType == "NC" {
		if err = db.Table("merchant_order_rate_list_view").Order("designation_no").Limit(1).Take(&resp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
			}
			return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
		}
	} else {
		if err = db.Table("merchant_order_rate_list_view").Find(&resp).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
		}
	}

	return resp, nil
}

func GetAllMerchantChannelRate(db *gorm.DB, merchantCode, currencyCode string) (resp []*types.MerchantOrderRateListViewX, err error) {
	//var terms []string
	//terms = append(terms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
	//terms = append(terms, "merchnrate_status = '1'")           // 0:禁用 1:啟用
	//terms = append(terms, "designation = '1'")
	//terms = append(terms, "chn_status = '1'")
	//terms = append(terms, "chnpaytype_status = '1'")
	//terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	//
	//term := strings.Join(terms, " AND ")
	db = db.Where("merchant_code = ?", merchantCode)
	db = db.Where("merchnrate_status = '1'")
	db = db.Where("designation = '1'")
	db = db.Where("chn_status = '1'")
	db = db.Where("chnpaytype_status = '1'")
	db = db.Where("currency_code = ?", currencyCode)

	if err = db.Table("merchant_order_rate_list_view").Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
	}

	return resp, nil
}

func GetMerchantChannelRateByCode(db *gorm.DB, merchantCode, currencyCode, channelCode string) (resp *types.MerchantOrderRateListViewX, err error) {
	//var merchantOrderRateListViews []*types.MerchantOrderRateListViewX
	//var terms []string
	//terms = append(terms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
	//terms = append(terms, "merchnrate_status = '1'")           // 0:禁用 1:啟用
	//terms = append(terms, fmt.Sprintf("pay_type_code = 'DF'")) // 內充看代付
	//terms = append(terms, "designation = '1'")
	//terms = append(terms, "chn_status = '1'")
	//terms = append(terms, "chnpaytype_status = '1'")
	//terms = append(terms, fmt.Sprintf("currency_code = '%s'", currencyCode))
	//
	//term := strings.Join(terms, " AND ")
	db = db.Where("merchant_code = ?", merchantCode)
	db = db.Where("merchnrate_status = '1'")
	db = db.Where("designation = '1'")
	db = db.Where("chn_status = '1'")
	db = db.Where("chnpaytype_status = '1'")
	db = db.Where("currency_code = ?", currencyCode)
	db = db.Where("pay_type_code = 'DF'")
	db = db.Where("channel_code = ?", channelCode)

	// 查询商户拥有的代付渠道
	if err = db.Table("merchant_order_rate_list_view").Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, "数据库错误: "+err.Error())
	}

	return resp, nil
}
