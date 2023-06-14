package permit

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitQueryAllLogic {
	return PermitQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitQueryAllLogic) PermitQueryAll(req types.PermitQueryAllRequest) (resp *types.PermitQueryAllResponse, err error) {
	var permits []types.Permit
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	if req.MenuID > 0 {
		//terms = append(terms, fmt.Sprintf("menu_id = '%d'", req.MenuID))
		db = db.Where("menu_id = ?", req.MenuID)
	}
	//term := strings.Join(terms, " AND ")
	db.Table("au_permits").Count(&count)
	err = db.Table("au_permits").Scopes(gormx.Paginate(req)).Find(&permits).Error

	resp = &types.PermitQueryAllResponse{
		List:     permits,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
