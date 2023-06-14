package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"gorm.io/gorm"
)

type DownReport struct {
	MyDB  *gorm.DB
	Table string
}

func NewDownReport(mydb *gorm.DB, t ...string) *DownReport {
	table := "rp_down_report"
	if len(t) > 0 {
		table = t[0]
	}
	return &DownReport{
		MyDB:  mydb,
		Table: table,
	}
}

func (d *DownReport) CreateReport(req *types.DownloadReportCreate) (err error) {
	return d.MyDB.Transaction(func(db *gorm.DB) error {
		return d.MyDB.Table(d.Table).Create(&req).Error
	})
}

func (d *DownReport) UpdateFailReport(id int64) (err error) {
	return d.MyDB.Transaction(func(db *gorm.DB) error {
		var downReportUp types.DownloadReportUpdate
		downReportUp.ID = id
		downReportUp.Status = constants.DOWN_FAIL
		return d.MyDB.Table(d.Table).Updates(downReportUp).Error
	})
}

func (d *DownReport) UpdateFinishReport(id int64, ss string) (err error) {
	return d.MyDB.Transaction(func(db *gorm.DB) error {
		var downReportUp types.DownloadReportUpdate
		downReportUp.ID = id
		downReportUp.Status = constants.DOWN_FINISH
		downReportUp.FilePath = ss
		return d.MyDB.Table(d.Table).Updates(downReportUp).Error
	})
}
