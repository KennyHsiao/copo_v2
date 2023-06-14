package model

import (
	"com.copo/bo_service/boadmin/internal/config"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
	"regexp"
	"strings"
)

type Channel struct {
	Config config.Config
	DB     *gorm.DB
	Table  string
}

func NewChannel(mydb *gorm.DB, t ...string) *Channel {
	table := "ch_channels"
	if len(t) > 0 {
		table = t[0]
	}
	return &Channel{
		DB:    mydb,
		Table: table,
	}
}

//检查渠道是否存在
func (m *Channel) CheckChannelIsExist(channelCode string) bool {
	var channel = &types.ChannelDataCreate{}
	m.DB.Table(m.Table).Where("code = ?", channelCode).Find(&channel)
	if channel == nil {
		return false
	}
	return true
}

//檢查白名單是否重複
func CheckIpDuplicated(whileList string) bool {
	ipList := strings.Split(whileList, ",")

	tempMap := make(map[string]int)

	for _, value := range ipList {
		tempMap[value] = 1
	}

	var keys []interface{}
	for k := range tempMap {
		keys = append(keys, k)
	}

	if len(keys) != len(ipList) {
		return true
	}
	return false
}

//檢查白名單是否格式錯誤 是:true 否 false
func CheckIpFormat(whiteList string) bool {
	var ipList []string
	ipList = append(ipList, strings.Split(whiteList, ",")...)
	for _, ip := range ipList {
		if isMatch, _ := regexp.MatchString(constants.RegexpIpaddressPattern, ip); !isMatch {
			return true
		}
	}
	return false
}

func CheckRequestCreateValue(req1 types.ChannelDataCreateRequest) error {

	if len(req1.WhiteList) > 0 {
		isIpFormateValid := CheckIpFormat(req1.WhiteList)
		if isIpFormateValid {
			return errorz.New(response.INVALID_WHITE_LIST)
		}

		isDuplicated := CheckIpDuplicated(req1.WhiteList)
		if isDuplicated {
			return errorz.New(response.WHITE_LIST_DUPLICATED)
		}
	}

	return nil
}

func CheckRequestUpdateValue(req1 types.ChannelDataUpdateRequest) error {

	if len(req1.WhiteList) > 0 {
		isIpFormateValid := CheckIpFormat(req1.WhiteList)
		if isIpFormateValid {
			return errorz.New(response.INVALID_WHITE_LIST)
		}

		isDuplicated := CheckIpDuplicated(req1.WhiteList)
		if isDuplicated {
			return errorz.New(response.WHITE_LIST_DUPLICATED)
		}
	}

	return nil
}
