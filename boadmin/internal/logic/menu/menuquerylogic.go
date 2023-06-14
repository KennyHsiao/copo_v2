package menu

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MenuQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MenuQueryLogic {
	return MenuQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuQueryLogic) MenuQuery(req types.MenuQueryRequest) (resp *types.MenuQueryResponse, err error) {

	err = l.svcCtx.MyDB.Table("au_menus").Preload("Permits").Take(&resp, req.ID).Error

	return
}
