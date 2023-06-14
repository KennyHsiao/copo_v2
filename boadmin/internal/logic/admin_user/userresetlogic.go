package admin_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserResetLogic {
	return UserResetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserResetLogic) UserReset(req types.UserResetRequest) (resp *types.UserResetResponse, err error) {

	if req.Target == "otp" {

		if err := l.svcCtx.MyDB.Table("au_users").
			Where("account = ?", req.Account).
			Updates(map[string]interface{}{
				"otp_key": "",
				"is_bind": "0",
			}).Error; err != nil {
			return nil, errors.New(response.DATABASE_FAILURE)
		}

	} else if req.Target == "password" {

		//ori := random.GetRandomString(10, random.ALL, random.MIX)
		ori := "00000000" // 管理員改密碼 預設八個零
		password := utils.PasswordHash2(ori)

		user := types.User{}
		res := l.svcCtx.MyDB.Table("au_users").
			Where("account = ?", req.Account).
			Updates(map[string]interface{}{
				"password": password,
				"is_login": "0",
			}).Scan(&user)

		if res.Error != nil {
			return nil, errors.New(response.DATABASE_FAILURE)
		}

		fmt.Println(">>>>>>>>>", user.Email, ori)

		//// 寄信
		//msg := gomail.NewMessage()
		//msg.SetHeader("From", "copoepay@copoonline.com")
		//msg.SetHeader("To", user.Email)
		//msg.SetHeader("Subject", "Reset Password!")
		//msg.SetBody("text/html", fmt.Sprintf("<b>Your New Password:</b> <br> %s", ori))
		//
		//go func() {
		//	if err := l.svcCtx.MailService.DialAndSend(msg); err != nil {
		//		logx.Error(">>>>>>>", err)
		//	}
		//}()

	} else {
		err := l.svcCtx.MyDB.Table("au_users").
			Where("account = ?", req.Account).
			Updates(map[string]interface{}{
				"is_login": "0",
			}).Error

		if err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	return nil, nil
}
