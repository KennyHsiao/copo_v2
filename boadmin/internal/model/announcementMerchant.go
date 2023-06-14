package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"gorm.io/gorm"
)

type AnnouncementMerchant struct {
	MyDB  *gorm.DB
	Table string
}

func NewAnnouncementMerchant(mydb *gorm.DB, t ...string) *AnnouncementMerchant {
	table := "an_announcement_merchants"
	if len(t) > 0 {
		table = t[0]
	}
	return &AnnouncementMerchant{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *AnnouncementMerchant) GetAnnouncementMerchant(announcementId int64, merchantCode string) (announcementMerchant *types.AnnouncementMerchant, err error) {
	err = m.MyDB.Table(m.Table).
		Where("announcement_id = ?", announcementId).
		Where("merchant_code = ?", merchantCode).
		Preload("Merchant").
		Take(&announcementMerchant).Error
	return
}

func (m *AnnouncementMerchant) FindByAnnouncementId(announcementId int64) (announcementMerchants []types.AnnouncementMerchant, err error) {
	err = m.MyDB.Table(m.Table).
		Where("announcement_id = ?", announcementId).
		Preload("Merchant").
		Find(&announcementMerchants).Error
	return
}

func (m *AnnouncementMerchant) FindByAnnouncementIdAndStatus(announcementId int64, status string) (announcementMerchants []types.AnnouncementMerchant, err error) {
	err = m.MyDB.Table(m.Table).
		Where("announcement_id = ?", announcementId).
		Where("status = ?", status).
		Preload("Merchant").
		Find(&announcementMerchants).Error
	return
}
