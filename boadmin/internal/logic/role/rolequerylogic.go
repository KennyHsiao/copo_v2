package role

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RoleQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) RoleQueryLogic {
	return RoleQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleQueryLogic) RoleQuery(req types.RoleQueryRequest) (resp *types.RoleQueryResponseX, err error) {
	err = l.svcCtx.MyDB.Table("au_roles").Preload("Menus", func(db *gorm.DB) *gorm.DB {
		return db.Order("au_menus.parent_id, au_menus.sort_order")
	}).Preload("Menus.Permits").Preload("Permits").Take(&resp, req.ID).Error

	resp.MenuTree, _ = json.Marshal(types.GenMenuTree(resp.Role.Menus))

	return
}
