package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

type MerchantChannelRate struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchantChannelRate(mydb *gorm.DB, t ...string) *MerchantChannelRate {
	table := "mc_merchant_channel_rate"
	if len(t) > 0 {
		table = t[0]
	}
	return &MerchantChannelRate{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *MerchantChannelRate) GetByMerchantCodeAndChannelPayTypeCode(merchantCode, channelPayTypeCode string) (merchantChannelRate *types.MerchantChannelRate, err error) {
	err = m.MyDB.Table(m.Table).
		Where("merchant_code = ?", merchantCode).
		Where("channel_pay_types_code = ? ", channelPayTypeCode).
		Take(&merchantChannelRate).Error
	return
}

func (m *MerchantChannelRate) GetMinMerChnFeeByPayTypeCode(channelPayTypeCode string) (merchantChannelRate *types.MerchantChannelRate, err error) {
	err = m.MyDB.Table("mc_merchant_channel_rate AS mmcr ").
		Joins("LEFT JOIN mc_merchants AS mm ON mmcr.merchant_code = mm.code ").
		Where("mmcr.channel_pay_types_code = ? ", channelPayTypeCode).
		Where("mm.rate_check != '0' ").
		Order("fee ASC").
		Limit(1).
		Take(&merchantChannelRate).Error
	return
}

func (m *MerchantChannelRate) GetMinMerChnHandlingFeeByPayTypeCode(channelPayTypeCode string) (merchantChannelRate *types.MerchantChannelRate, err error) {
	err = m.MyDB.Table("mc_merchant_channel_rate AS mmcr ").
		Joins("LEFT JOIN mc_merchants AS mm ON mmcr.merchant_code = mm.code ").
		Where("mmcr.channel_pay_types_code = ? ", channelPayTypeCode).
		Where("mm.rate_check != '0' ").
		Order("handling_fee ASC").
		Limit(1).
		Take(&merchantChannelRate).Error
	return
}

func (m *MerchantChannelRate) DeleteByMerchantCodeAndChannelPayTypeCode(merchantCode, channelPayTypeCode string) (err error) {
	err = m.MyDB.Table(m.Table).
		Where("merchant_code = ?", merchantCode).
		Where("channel_pay_types_code = ?", channelPayTypeCode).Delete(&types.MerchantChannelRate{}).Error
	return
}
