package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
)

type orderAction struct {
	MyDB  *gorm.DB
	Table string
}

func NewOrderAction(mydb *gorm.DB, t ...string) *orderAction {
	table := "tx_order_actions"
	if len(t) > 0 {
		table = t[0]
	}
	return &orderAction{
		MyDB:  mydb,
		Table: table,
	}
}

func (u *orderAction) CreateOrderAction(orderAction *types.OrderActionX) error {
	if err := u.MyDB.Table(u.Table).Create(orderAction).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
