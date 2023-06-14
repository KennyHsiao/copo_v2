package admin_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserQueryAllLogic {
	return UserQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserQueryAllLogic) UserQueryAll(req types.UserQueryAllRequest) (resp *types.UserQueryAllResponse, err error) {
	var users []types.User
	var count int64

	db := l.svcCtx.MyDB.Table("au_users")

	if len(req.Account) > 0 {
		db = db.Where("au_users.account like ?", "%"+req.Account+"%")
	}
	if len(req.Name) > 0 {
		db = db.Where("au_users.name like ?", "%"+req.Name+"%")
	}
	if len(req.Email) > 0 {
		db = db.Where(" au_users.`email` like ?", "%"+req.Email+"%")
	}
	if len(req.Status) > 0 {
		db = db.Where(" au_users.`status` = ?", req.Status)
	}
	if len(req.RoleName) > 0 {
		db = db.Joins("join au_user_roles on user_id = au_users.id ").
			Joins("join au_roles on au_roles.id = role_id and au_roles.name = ?", req.RoleName).
			Group("au_users.id")
	}
	db = db.Table("au_users").Where("au_users.`is_admin` = ?", constants.IS_ADMIN_YES)

	err = db.Count(&count).Error

	err = db.
		Preload("Roles.Menus.Permits").
		Preload("Roles.Permits").
		Preload("Merchants").
		Scopes(gormx.Paginate(req)).Find(&users).Error

	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.UserQueryAllResponse{
		List:     users,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
