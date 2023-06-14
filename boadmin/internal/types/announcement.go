package types

import (
	"com.copo/bo_service/common/gormx"
	"time"
)

func (ChannelRule) TableName() string {
	return "an_channel_rule"
}

func (AnnouncementChannel) TableName() string {
	return "an_announcement_channels"
}

func (AnnouncementMerchant) TableName() string {
	return "an_announcement_merchants"
}

func (AnnouncementParam) TableName() string {
	return "an_announcement_params"
}

type AnnouncementTempQueryAllRequestX struct {
	AnnouncementTempQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type AnnouncementTempParamQueryAllRequestX struct {
	AnnouncementTempParamQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type AnnouncementQueryAllRequestX struct {
	AnnouncementQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type AnnouncementX struct {
	Announcement
	CreatedAt JsonTime `json:"createdAt"`
	UpdatedAt JsonTime `json:"updatedAt"`
}

type AnnouncementQueryAllResponseX struct {
	List     []AnnouncementX `json:"list"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	RowCount int64           `json:"rowCount"`
}

type ChannelRuleCreate struct {
	ChannelRuleCreateRequest
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChannelRuleUpdate struct {
	ChannelRuleUpdateRequest
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChannelRuleQueryRequestX struct {
	ChannelRuleQueryRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type AnnouncementCreate struct {
	AnnouncementCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementUpdate struct {
	AnnouncementUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementUpdateDraft struct {
	AnnouncementUpdateDraftRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementChannelX struct {
	AnnouncementChannel
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementMerchantX struct {
	AnnouncementMerchant
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AnnouncementParamX struct {
	AnnouncementParam
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChannelRuleQueryX struct {
	ChannelRuleQuery
	UpdatedAt JsonTime `json:"updatedAt"`
}

type ChannelRuleQueryResponseX struct {
	List     []ChannelRuleQueryX `json:"list"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
	RowCount int64               `json:"rowCount"`
}

type GroupQueryX struct {
	GroupQueryResponse
	CreatedAt JsonTime `json:"createdAt"`
	UpdatedAt JsonTime `json:"updatedAt"`
}

type GroupQueryResponseX struct {
	List     []GroupQueryX `json:"list"`
	PageNum  int           `json:"pageNum"`
	PageSize int           `json:"pageSize"`
	RowCount int64         `json:"rowCount"`
}
