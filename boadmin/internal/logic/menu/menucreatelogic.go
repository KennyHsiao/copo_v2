package menu

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type MenuCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MenuCreateLogic {
	return MenuCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuCreateLogic) MenuCreate(req types.MenuCreateRequest) error {
	menu := &types.MenuCreate{
		MenuCreateRequest: req,
	}
	var ignoredFields []string
	if req.ParentID == 0 {
		ignoredFields = append(ignoredFields, "parent_id")
	}

	return l.svcCtx.MyDB.Table("au_menus").Omit(ignoredFields...).Create(menu).Error
}
