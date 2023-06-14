package menu

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type MenuQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMenuQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MenuQueryAllLogic {
	return MenuQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MenuQueryAllLogic) MenuQueryAll(req types.MenuQueryAllRequestX) (resp *types.MenuQueryAllResponse, err error) {
	var menus []types.Menu
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("au_menus")

	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name like '%%%s%%'", req.Name))
		db = db.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.Group) > 0 {
		//terms = append(terms, fmt.Sprintf("`group` like '%%%s%%'", req.Group))
		db = db.Where("`group` like ?", "%"+req.Group+"%")
	}
	//term := strings.Join(terms, " AND ")
	db.Count(&count)
	err = db.
		Preload("Permits").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&menus).Error

	resp = &types.MenuQueryAllResponse{
		List:     menus,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
