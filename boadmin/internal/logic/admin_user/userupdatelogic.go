package admin_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type UserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserUpdateLogic {
	return UserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUpdateLogic) UserUpdate(req types.UserUpdateRequest) (err error) {
	var user types.User
	l.svcCtx.MyDB.Table("au_users").Find(&user, map[string]interface{}{
		"id": req.ID,
	})

	if len(req.OldPassword) > 0 {
		if !utils.CheckPassword2(req.OldPassword, user.Password) {
			return errorz.New(response.INCORRECT_USER_PASSWD, "密碼錯誤")
		}
		req.Password = utils.PasswordHash2(req.Password)
	}

	update := &types.UserUpdate{
		UserUpdateRequest: req,
	}
	// 清除關聯
	_ = l.svcCtx.MyDB.Model(&user).Association("Roles").Clear()
	_ = l.svcCtx.MyDB.Model(&user).Association("Currencies").Clear()
	_ = l.svcCtx.MyDB.Model(&user).Association("Email").Clear()

	err = l.svcCtx.MyDB.Table("au_users").Where("account = ?", req.Account).
		Updates(map[string]interface{}{"email": req.Email}).Error

	// Omit("Roles.*") 關閉自動維護upsert功能
	err = l.svcCtx.MyDB.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Updates(update).Error

	return err
}
