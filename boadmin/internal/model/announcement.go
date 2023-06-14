package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

type Announcement struct {
	MyDB  *gorm.DB
	Table string
}

func NewAnnouncement(mydb *gorm.DB, t ...string) *Announcement {
	table := "an_announcements"
	if len(t) > 0 {
		table = t[0]
	}
	return &Announcement{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *Announcement) GetAnnouncement(id int64) (announcement *types.Announcement, err error) {
	err = m.MyDB.Table(m.Table).
		Preload("AnnouncementChannels").
		Preload("AnnouncementMerchants.Merchant").
		Preload("AnnouncementParams").
		Take(&announcement, id).Error
	return
}

func (m *Announcement) AutoChangeStatus(id int64) (err error) {
	var announcementMerchants []types.AnnouncementMerchant

	if err = m.MyDB.Table("an_announcement_merchants").
		Where("announcement_id = ?", id).
		Find(&announcementMerchants).Error; err != nil {
		return err
	}

	isHasDraft := false
	isHasSuccess := false
	isHasFailure := false
	isHasUnsend := false
	isHasUnsendFailure := false

	// 1=草稿/2=成功/3=失敗/4=回收/5=回收失敗/6=忽略
	for _, merchant := range announcementMerchants {
		switch merchant.Status {
		case "1":
			isHasDraft = true
		case "2":
			isHasSuccess = true
		case "3":
			isHasFailure = true
		case "4":
			isHasUnsend = true
		case "5":
			isHasUnsendFailure = true
		}
	}
	// 預設失敗
	status := 3
	if isHasUnsendFailure {
		// 只要有回收失敗 = 回收失敗
		status = 5
	} else if isHasFailure {
		// 只要有失敗(沒回收失敗) = 失敗
		status = 3
	} else if !isHasDraft && !isHasUnsend {
		// 沒草稿 回收 = 成功
		status = 2
	} else if !isHasDraft && !isHasSuccess {
		// 沒草稿 成功 = 回收
		status = 4
	}

	return m.MyDB.Table(m.Table).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
		}).Error
}
