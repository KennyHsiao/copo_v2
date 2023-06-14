package role

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type RoleQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRoleQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) RoleQueryAllLogic {
	return RoleQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RoleQueryAllLogic) RoleQueryAll(req types.RoleQueryAllRequest) (resp *types.RoleQueryAllResponse, err error) {
	var roles []types.Role
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	db2 := l.svcCtx.MyDB
	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name = '%%%s%%'", req.Name))
		db = db.Where("name like ?", "%"+req.Name+"%")
		db2 = db2.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("type = '%s'", req.Type))
		db = db.Where("type = ?", req.Type)
		db2 = db2.Where("name like ?", "%"+req.Name+"%")
	}
	//term := strings.Join(terms, " AND ")
	db2.Table("au_roles").Count(&count)
	err = db.Select("au_roles.*, name_i18n->>'$." + req.Language + "' as name").Table("au_roles").Preload("Menus.Permits").Preload("Permits").Scopes(gormx.Paginate(req)).Find(&roles).Error

	resp = &types.RoleQueryAllResponse{
		List:     roles,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
