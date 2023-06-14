package channelDataService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"errors"
	"gorm.io/gorm"
)

func Query_N_ChannelData(db *gorm.DB, channelCodeList []string) (*[]types.ChannelDataUpdate, error) {

	channelDataList := &[]types.ChannelDataUpdate{}
	if err := db.Table("ch_channels").Where("code IN ? ", channelCodeList).Find(channelDataList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.CHANNEL_IS_NOT_EXIST, err.Error())
		} else if err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}
	return channelDataList, nil

}
