package role

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type RoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) RoleUpdateLogic {
	return RoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleUpdateLogic) RoleUpdate(req types.RoleUpdateRequest) error {
	role := &types.RoleUpdate{
		RoleUpdateRequest: req,
	}

	l.svcCtx.MyDB.Exec("DELETE FROM au_role_menus WHERE role_id = ?", req.ID)
	l.svcCtx.MyDB.Exec("DELETE FROM au_role_permits WHERE role_id = ?", req.ID)
	return l.svcCtx.MyDB.Table("au_roles").Updates(role).Error
}
