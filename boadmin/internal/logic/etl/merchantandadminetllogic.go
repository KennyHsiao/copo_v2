package etl

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"context"
	"encoding/json"
	"github.com/copo888/transaction_service/common/response"
	"gorm.io/gorm"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantAndAdminEtlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type account struct {
	AccountName      string `json:"account_name"`
	Password         string `json:"password"`
	VerificationKey  string `json:"verification_key"`
	MerchantCoding   string `json:"merchant_coding"`
	FailCount        string `json:"fail_count"`
	ActivationStatus string `json:"activation_status"`
	LastLoginAt      string `json:"last_login_at"`
	SystemId         string `json:"system_id"`
	Status           string `json:"status"`
	Currency         string `json:"currency"`
}

type UserCurrency struct {
	UserAccount  string `json:"user_account"`
	CurrencyCode string `json:"currency_code"`
}

type UserRole struct {
	UserId int64 `json:"user_id"`
	RoleId int64 `json:"role_id"`
}

func NewMerchantAndAdminEtlLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantAndAdminEtlLogic {
	return MerchantAndAdminEtlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantAndAdminEtlLogic) MerchantAndAdminEtl(req *types.MerchantAndAdminInfoRequest) error {
	//oldDb, err := mysqlz.New("8.129.209.41", "3306", "dior", "P#tjnnPEZ@JwQjkFrcdG", "dior22").
	//	SetCharset("utf8mb4").
	//	SetLoc("UTC").
	//	SetLogger(logrusz.New().SetLevel("debug").Writer()).
	//	Connect(mysqlz.Pool(1, 1, 1))
	//if err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, err.Error())
	//}
	//// 帳號與帳號幣別導入
	//if err := l.AccountAndAcoountCurrencies(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "帳號與帳號幣別導入錯誤："+err.Error())
	//}
	//// 管理者權限導入
	//if err := l.AccountRole(); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "管理者權限導入錯誤："+err.Error())
	//}
	// 	更新使用者啟用狀態
	//if err := l.UpdateEnableOrDisable(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "更新使用者啟用狀態錯誤："+err.Error())
	//}
	return nil
}

func (l *MerchantAndAdminEtlLogic) AccountAndAcoountCurrencies(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var accounts []account
		if err := oldDb.Table("account").Find(&accounts).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var userXs []types.UserX
		var userCurrenies []UserCurrency
		for _, record := range accounts {
			isLogin := "0"
			if record.Status == "verifying" {
				isLogin = "1"
			} else if record.Status == "normal" {
				isLogin = "2"
			}
			isAdmin := "0"

			if record.SystemId == "009" {
				isAdmin = "1"
			}
			isFreeze := "0"
			if record.Status == "freeze" {
				isFreeze = "1"
			}
			disableDelete := "0"
			if isAdmin == "1" {
				disableDelete = "1"
			}
			activationStatus := "1"
			if record.ActivationStatus != "1" {
				activationStatus = "0"
			}
			regTime := time.Now().UTC().Unix()
			if len(record.VerificationKey) > 0 {
				user := types.User{
					Account:       record.AccountName,
					Name:          record.AccountName,
					Password:      record.Password,
					IsLogin:       isLogin,
					IsAdmin:       isAdmin,
					OtpKey:        record.VerificationKey,
					Qrcode:        "IsV1Data",
					Status:        activationStatus,
					FailCount:     record.FailCount,
					IsFreeze:      isFreeze,
					IsBind:        "1",
					DisableDelete: disableDelete,
					RegisteredAt:  regTime,
				}

				userX := types.UserX{
					User: user,
				}

				userXs = append(userXs, userX)
			} else {
				user := types.User{
					Account:       record.AccountName,
					Name:          record.AccountName,
					Password:      record.Password,
					IsLogin:       isLogin,
					IsAdmin:       isAdmin,
					OtpKey:        record.VerificationKey,
					Status:        activationStatus,
					FailCount:     record.FailCount,
					IsFreeze:      isFreeze,
					IsBind:        "0",
					DisableDelete: disableDelete,
					RegisteredAt:  regTime,
				}

				userX := types.UserX{
					User: user,
				}

				userXs = append(userXs, userX)
			}

			//if err := tx.Table("au_users_copy1").Create(&userX).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}
			var arr []string
			_ = json.Unmarshal([]byte(record.Currency), &arr)

			for _, r := range arr {
				userCurrency := UserCurrency{
					UserAccount:  record.AccountName,
					CurrencyCode: r,
				}

				userCurrenies = append(userCurrenies, userCurrency)
			}

		}
		if err := tx.Table("au_users").CreateInBatches(&userXs, len(userXs)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err := tx.Table("au_user_currencies").CreateInBatches(userCurrenies, len(userCurrenies)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantAndAdminEtlLogic) AccountRole() error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var users []types.User
		if err := tx.Table("au_users").Where("is_admin = '1'").Find(&users).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var userRoles []UserRole
		for _, user := range users {
			userRole := UserRole{
				UserId: user.ID,
				RoleId: 72,
			}
			userRoles = append(userRoles, userRole)
		}
		if err := l.svcCtx.MyDB.Table("au_user_roles").CreateInBatches(userRoles, len(userRoles)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantAndAdminEtlLogic) UpdateEnableOrDisable(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var accounts []account
		if err := oldDb.Table("account").Find(&accounts).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var userXs []types.UserX
		if err := tx.Table("au_users").Find(&userXs).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		userXMap := make(map[string]types.UserX)
		for _, x := range userXs {
			userXMap[x.Account] = x
		}
		for _, a := range accounts {
			activationStatus := "1"
			if a.ActivationStatus != "1" {
				activationStatus = "0"
			}
			if v, ok := userXMap[a.AccountName]; ok {
				v.Status = activationStatus
				if err := tx.Table("au_users").Updates(v).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			}
		}

		return nil
	})
}
