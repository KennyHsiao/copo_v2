package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

type ChannelBank struct {
	MyDB  *gorm.DB
	Table string
}

func NewChannelBank(mydb *gorm.DB, t ...string) *ChannelBank {
	table := "ch_channel_banks"
	if len(t) > 0 {
		table = t[0]
	}
	return &ChannelBank{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *ChannelBank) InsertChannelBank(req []types.ChannelBankCreateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		return m.MyDB.Table(m.Table).Create(req).Error
	})
}

func (m *ChannelBank) UpdateChannelBank(req []types.ChannelBankUpdateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		for _, r := range req {
			err = m.MyDB.Table(m.Table).Updates(r).Error
		}
		return err
	})
}
