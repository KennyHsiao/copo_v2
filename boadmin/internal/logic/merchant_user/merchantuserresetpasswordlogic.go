package merchant_user

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUserResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUserResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUserResetPasswordLogic {
	return MerchantUserResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUserResetPasswordLogic) MerchantUserResetPassword(req types.MerchantUserResetPasswordRequest) (err error) {
	password := random.GetRandomString(10, random.ALL, random.MIX)
	var user types.User

	return l.svcCtx.MyDB.Transaction(func(txDB *gorm.DB) (err error) {

		if err = txDB.Table("au_users").Where("account = ?", req.Account).Take(&user).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		user.Password = utils.PasswordHash2(password)
		user.IsLogin = "0"
		if err = txDB.Table("au_users").
			Omit("Merchants.*").
			Omit("Roles.*").Updates(&types.UserX{User: user}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 寄信
		body := "<table>" +
			"    <thead>" +
			"    <span>Dear " + req.Account + " 用戶您好</span>" +
			"    </thead>" +
			"    <tbody>" +
			"    <div>" +
			"         <span>你的密码已重置，你的预设密码如下</span><br>" +
			"    </div>" +
			"    <div>" +
			"        <span>密碼 password:<span>" + password + "</span><br>" +
			"    </div>" +
			"    <div>" +
			"        <span>再麻烦至本系统作登录设定 ，谢谢!</span></span><br>" +
			"    </div>" +
			"    </tbody>" +
			"</table>"

		if err = utils.SendEmail(l.svcCtx.MailService, l.svcCtx.Config.Smtp.User, user.Email, "密码重置通知", body); err != nil {
			return err
		}
		return
	})
}
