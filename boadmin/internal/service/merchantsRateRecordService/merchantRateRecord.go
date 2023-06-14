package merchantsRateRecordService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

func CreateMerchantRateRecord(db *gorm.DB, req *types.MerchantRateRecordCreateRequest) error {
	merchantRateRecordCreate := &types.MerchantRateRecordCreate{
		MerchantRateRecordCreateRequest: *req,
	}
	return db.Transaction(func(db *gorm.DB) error {
		return db.Table("mc_merchant_rate_record").Create(merchantRateRecordCreate).Error
	})
}
