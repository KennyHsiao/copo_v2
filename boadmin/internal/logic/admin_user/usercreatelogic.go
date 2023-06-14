package admin_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/utils"
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserCreateLogic {
	return UserCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCreateLogic) UserCreate(req types.UserCreateRequest) error {

	req.Password = utils.PasswordHash2(req.Password)
	// 管理員不需改密碼
	req.IsLogin = "1"
	req.IsAdmin = "1"
	req.RegisteredAt = time.Now().Unix()

	user := &types.UserCreate{
		UserCreateRequest: req,
	}

	return l.svcCtx.MyDB.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Create(user).Error
}
