package channelRateRecordService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

func CreateChannelRateRecord(db *gorm.DB, req *types.ChannelRateRecordCreateRequest) error {
	channelRateRecordCreate := &types.ChannelRateRecordCreate{
		ChannelRateRecordCreateRequest: *req,
	}

	return db.Transaction(func(db *gorm.DB) error {
		return db.Table("ch_channel_rate_record").Create(channelRateRecordCreate).Error
	})
}
