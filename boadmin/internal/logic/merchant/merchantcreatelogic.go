package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

type MerchantCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantCreateLogic {
	return MerchantCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantCreateLogic) MerchantCreate(req types.MerchantCreateRequest) error {
	merchant := &types.MerchantCreate{
		MerchantCreateRequest: req,
	}
	setInitialValue(merchant)

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {

		if err := verify(db, merchant); err != nil {
			return err
		}

		mux := model.NewMerchant(db)
		merchant.Code = mux.GetNextMerchantCode()

		password := random.GetRandomString(10, random.ALL, random.MIX)
		user, err1 := userCreate(db, merchant, password)
		if err1 != nil {
			if strings.Contains(err1.Error(), "Duplicate entry") {
				fmt.Println("Duplicate entry error")
				return errorz.New(response.ACCOUNT_IS_ALREADY_IN_USE, err1.Error())
			} else {
				return errorz.New(response.DATABASE_FAILURE, err1.Error())
			}

		}

		mcux := model.NewMerchantCurrency(db)
		mbux := model.NewMerchantBalance(db)

		for _, currency := range req.MerchantCurrencies {
			if err := mcux.CreateMerchantCurrency(merchant.Code, currency.CurrencyCode, "1", currency.SortOrder); err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			if err := mbux.CreateMerchantBalances(merchant.Code, currency.CurrencyCode); err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		if err := db.Table("mc_merchants").
			Omit("Users.*").Omit("MerchantCurrencies").
			Create(merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 寄信
		body := "<table>" +
			"    <thead>" +
			"    <span>Dear 用戶您好</span>" +
			"    </thead>" +
			"    <tbody>" +
			"    <div>" +
			"        <span>感謝您使本系統</span><br>" +
			"    </div>" +
			"    <div>" +
			"        <span>登錄帳號 accountName:<span> " + user.Account + " </span><br>" +
			"    </div>" +
			"    <div>" +
			"        <span>密碼 password:<span>" + password + "</span><br>" +
			"    </div>" +
			"    <div>" +
			"        <span>再麻烦至本系统作登录设定 ，谢谢!</span></span><br>" +
			"    </div>" +
			"    </tbody>" +
			"</table>"

		if err := utils.SendEmail(l.svcCtx.MailService, l.svcCtx.Config.Smtp.User, user.Email, "注册结果通知", body); err != nil {
			return err
		}

		return nil
	})

}

func setInitialValue(merchant *types.MerchantCreate) {

	merchant.ScrectKey = random.GetRandomString(32, random.ALL, random.MIX)
	merchant.AccountName = merchant.Account
	merchant.Status = constants.MerchantStatusEnable
	merchant.AgentStatus = constants.MerchantAgentStatusDisable
	merchant.LoginValidatedType = "1"
	merchant.PayingValidatedType = "1"
	merchant.ApiCodeType = "1"
	merchant.BillLadingType = "0"
	merchant.RateCheck = "1"
	merchant.Lang = "CN"
	merchant.RegisteredAt = time.Now().Unix()
	merchant.Users = append(merchant.Users, types.User{Account: merchant.Account})
}

func verify(db *gorm.DB, merchant *types.MerchantCreate) (err error) {
	var isExist bool

	if isExist, err = model.NewUser(db).IsExistByAccount(merchant.Account); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if isExist {
		return errorz.New(response.USER_HAS_REGISTERED, "该用户名已存在")
	}

	//if isExist, err = model.NewUser(db).IsExistByEmail(merchant.Contact.Email); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, err.Error())
	//} else if isExist {
	//	return errorz.New(response.MAILBOX_HAS_REGISTERED, "该邮箱已存在")
	//}

	var ips []string

	if len(merchant.BoIP) > 0 {
		ips = append(ips, strings.Split(merchant.BoIP, ",")...)
	}
	if len(merchant.ApiIP) > 0 {
		ips = append(ips, strings.Split(merchant.ApiIP, ",")...)
	}

	for _, ip := range ips {
		if isMatch, _ := regexp.MatchString(constants.RegexpIpaddressPattern, ip); !isMatch {
			return errorz.New(response.ILLEGAL_IP, "IP格式错误")
		}
	}

	return
}

func userCreate(db *gorm.DB, merchant *types.MerchantCreate, password string) (*types.UserCreate, error) {

	var currencies []types.Currency

	for _, currency := range merchant.MerchantCurrencies {
		currencies = append(currencies, types.Currency{
			Code: currency.CurrencyCode,
		})
	}
	user := &types.UserCreate{
		UserCreateRequest: types.UserCreateRequest{
			Account:      merchant.Account,
			Name:         merchant.Account,
			Email:        merchant.Contact.Email,
			RegisteredAt: time.Now().Unix(),
			Roles: []types.Role{{
				ID: 2,
			}},
			Password:      utils.PasswordHash2(password),
			Currencies:    currencies,
			DisableDelete: "1",
			Status:        "1",
			IsLogin:       "0",
			IsAdmin:       "0",
		},
	}

	return user, db.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Create(user).Error
}
