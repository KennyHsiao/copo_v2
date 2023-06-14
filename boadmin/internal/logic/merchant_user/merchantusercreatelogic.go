package merchant_user

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUserCreateLogic {
	return MerchantUserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUserCreateLogic) MerchantUserCreate(req types.MerchantUserCreateRequest) (err error) {
	password := random.GetRandomString(10, random.ALL, random.MIX)
	user := types.User{
		Account:       req.Account,
		Name:          req.Account,
		Phone:         req.Phone,
		Email:         req.Email,
		RegisteredAt:  time.Now().Unix(),
		Roles:         req.Roles,
		Merchants:     req.Merchants,
		Password:      utils.PasswordHash2(password),
		Currencies:    req.Currencies,
		DisableDelete: "0",
		Status:        "1",
		IsLogin:       "0",
		IsAdmin:       "0",
		IsFreeze:      "0",
	}

	if err = l.svcCtx.MyDB.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Create(&types.UserX{User: user}).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			fmt.Println("Duplicate entry error")
			return errorz.New(response.ACCOUNT_IS_ALREADY_IN_USE, err.Error())
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
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
		"        <span>密碼 password:<span> " + password + "</span><br>" +
		"    </div>" +
		"    <div>" +
		"        <span>再麻烦至本系统作登录设定 ，谢谢!</span></span><br>" +
		"    </div>" +
		"    </tbody>" +
		"</table>"

	if err := utils.SendEmail(l.svcCtx.MailService, l.svcCtx.Config.Smtp.User, user.Email, "注册结果通知", body); err != nil {
		return err
	}

	return
}
