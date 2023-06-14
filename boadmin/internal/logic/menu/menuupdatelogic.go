package menu

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MenuUpdateLogic {
	return MenuUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuUpdateLogic) MenuUpdate(req types.MenuUpdateRequest) error {
	menu := &types.MenuUpdate{
		MenuUpdateRequest: req,
	}
	return l.svcCtx.MyDB.Table("au_menus").Session(&gorm.Session{FullSaveAssociations: true}).Updates(menu).
		Update("hidden", menu.Hidden).Error
}
