package types

import (
	"com.copo/bo_service/common/gormx"
	"time"
)

func (Bank) TableName() string {
	return "bk_banks"
}

type BankBlockAccountCreate struct {
	BankBlockAccountCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy string
	UpdatedBy string
}

type BankBlockAccountUpdate struct {
	BankBlockAccountUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy string
}

type BankCreate struct {
	BankCreateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BankUpdate struct {
	BankUpdateRequest
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BankQueryAllRequestX struct {
	BankQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type BankQueryForBKRequestX struct {
	BankQueryForBKRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}
