package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

type user struct {
	MyDB  *gorm.DB
	Table string
}

func NewUser(mydb *gorm.DB, t ...string) *user {
	table := "au_users"
	if len(t) > 0 {
		table = t[0]
	}
	return &user{
		MyDB:  mydb,
		Table: table,
	}
}

func (u *user) GetUser(id int64) (resp *types.UserQueryResponse, err error) {
	err = u.MyDB.Table(u.Table).
		Preload("Roles.Menus.Permits").
		Preload("Roles.Permits").
		Preload("Merchants").
		Preload("Currencies").
		Take(&resp, id).Error
	return
}

func (u *user) IsExistByAccount(account string) (isExist bool, err error) {
	err = u.MyDB.Table(u.Table).
		Select("count(*) > 0").
		Where("account = ?", account).
		Find(&isExist).Error
	return
}

func (u *user) IsExistByEmail(email string) (isExist bool, err error) {
	err = u.MyDB.Table(u.Table).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&isExist).Error
	return
}

type UserPermit struct {
	PermitId int64
}

type Login struct {
	ID        int64
	Account   string
	Name      string
	Password  string
	OtpKey    string
	Qrcode    string
	IsLogin   string
	IsBind    string
	IsAdmin   string
	IsFreeze  string
	FailCount int64
	Status    string
	Roles     []types.Role     `json:"roles" gorm:"many2many:au_user_roles;foreignKey:id;joinForeignKey:user_id;"`
	Merchants []types.Merchant `json:"merchantsService" gorm:"many2many:au_user_merchants;foreignKey:account;joinForeignKey:user_account;references:code;joinReferences:merchant_code;"`
	MenuTree  types.MenuTree   `gorm:"-"`
}

func (o *Login) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), o)
}

func (o Login) Value() (driver.Value, error) {
	b, err := json.Marshal(o)
	return string(b), err
}

func (u *user) Menu(account string) (Login, error) {
	var member Login
	if err := u.MyDB.Table(u.Table).
		Where("account = ?", account).
		Preload("Roles.Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("au_menus.parent_id, au_menus.sort_order")
		}).
		Preload("Roles.Menus.Permits").
		Preload("Merchants").
		Take(&member).Error; err != nil {
		return Login{}, errors.New(response.USER_NOT_FOUND)
	}

	return member, nil
}

func (u *user) Auth(account string, pwd string, ctx context.Context) (Login, error) {

	var member Login
	db := u.MyDB.Table(u.Table).
		Where("account = ?", account).
		Preload("Roles.Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("au_menus.parent_id, au_menus.sort_order")
		}).
		Preload("Roles.Menus.Permits").
		Preload("Merchants").
		Take(&member)

	if db.Error != nil {
		logx.WithContext(ctx).Error(db.Error)
		return Login{}, errors.New(response.USER_CAN_NOT_LOGIN)
	}

	if member.Status != "1" {
		return Login{}, errors.New(response.USER_ACTIVATION_STATUS_DISABLE)
	}

	if member.IsFreeze == "1" {
		return Login{}, errors.New(response.USER_STATUS_FREEZE)
	}

	if member.IsAdmin == "0" && len(member.Merchants) < 1 {
		return Login{}, errors.New(response.INVALID_MERCHANT_CODING)
	}

	if member.IsAdmin == "0" && member.Merchants[0].Status == "0" {
		return Login{}, errors.New(response.MERCHANT_IS_DISABLE)
	}

	if member.IsAdmin == "0" && member.Merchants[0].Status == "0" {
		return Login{}, errors.New(response.MERCHANT_IS_DISABLE)
	}

	fmt.Println(">>>>", utils.CheckPassword2(pwd, member.Password))

	if utils.CheckPassword2(pwd, member.Password) {
		fmt.Println(">>>>", u.UpdateFailCount(member.Account, 0))
		return member, nil
	} else {
		fmt.Println(">>>>", u.UpdateFailCount(member.Account, member.FailCount+1))
	}

	return Login{}, errors.New(response.INCORRECT_USER_PASSWD)
}

func (u *user) Exists(account string) (Login, error) {

	var member Login
	db := u.MyDB.Table(u.Table).
		Where("account = ?", account).
		Take(&member)

	if db.Error == nil {
		return member, nil
	}

	return Login{}, errors.New(response.USER_NOT_FOUND)
}

func (u *user) UpdateFailCount(account string, failCount int64) error {
	isFreeze := 0
	if failCount == 5 {
		isFreeze = 1
	}
	return u.MyDB.Table(u.Table).Where("account = ?", account).
		Updates(map[string]interface{}{
			"fail_count": failCount,
			"is_freeze":  isFreeze,
		}).Error
}

func (u *user) UpdatelLastLogin(account string, ip string) error {
	return u.MyDB.Table(u.Table).Where("account = ?", account).
		Updates(map[string]interface{}{
			"last_login_at": time.Now().Unix(),
			"last_login_ip": ip,
		}).Error
}
