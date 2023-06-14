package merchantchannelrateservice

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"fmt"
	"gorm.io/gorm"
)

// @param merchantCode 商户号
// @param payTypeCode 支付类型代码
// @param currencyCode 币别
// @param payTypeNo 支付类型编码
// @param billLadingType 提單類型 (0=單指、1=多指)
// @return

func GetDesignationMerChnRate(db *gorm.DB, merchantCode, payTypeCode, currencyCode, payTypeNo, billLadingType string) (correspondMerChnRate *types.CorrespondMerChnRate, err error) {

	selectX := "mmcr.merchant_code," + //商户号
		"mmcr.channel_pay_types_code," + //渠道支付编码
		"mmcr.channel_code," + //渠道编号
		"mmcr.pay_type_code," + //支付类型代码
		"mmcr.designation," + //是否指定渠道
		"mmcr.designation_no," + //指定渠道编码
		"mmcr.status as merchant_status," + //状态
		"mmcr.fee," + //商户费率
		"mmcr.handling_fee," + //商户手续费
		"ccpt.map_code," + //渠道支付類型代碼
		"ccpt.fee as ch_fee," + //渠道费率
		"ccpt.handling_fee as ch_handling_fee," + //渠道手续费
		"ccpt.single_min_charge," + //
		"ccpt.single_max_charge," +
		"cc.currency_code," +
		"cc.channel_port," +
		"cc.api_url"

	db = db.Where(" mmcr.`designation` = '1'").
		Where("mmcr.`status` = '1'").
		Where("ccpt.`status` = '1'").
		Where("cc.`status` = '1'").
		Where("mmcr.`merchant_code` = ?", merchantCode).
		Where("mmcr.`pay_type_code` = ?", payTypeCode).
		Where("cc.`currency_code` = ?", currencyCode)

	if billLadingType == "1" {
		// 多指定要給指定代碼
		if payTypeNo != "" {
			db = db.Where("mmcr.`designation_no` = ?", payTypeNo)
		} else {
			return nil, errorz.New(response.NO_CHANNEL_SET)
		}
	}

	if err = db.Select(selectX).
		Table("mc_merchant_channel_rate as mmcr ").
		Joins("join ch_channels cc on mmcr.channel_code = cc.code ").
		Joins("join ch_channel_pay_types ccpt on mmcr.channel_pay_types_code = ccpt.code").
		Order("designation_no asc").
		Take(&correspondMerChnRate).Error; err != nil {
		return nil, errorz.New(response.SYSTEM_ERROR, fmt.Sprintf("商户代码[%s]或支付类型代码[%s]或幣別[%s]错误或指定渠道设定错误或关闭或维护", merchantCode, payTypeCode, currencyCode))
	}
	return
}
