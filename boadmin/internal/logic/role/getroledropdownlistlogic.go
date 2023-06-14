package role

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleDropDownListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleDropDownListLogic(ctx context.Context, svcCtx *svc.ServiceContext) GetRoleDropDownListLogic {
	return GetRoleDropDownListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleDropDownListLogic) GetRoleDropDownList(req *types.RoleDropDownItemRequest) (items []types.RoleDropDownItem, err error) {
	db := l.svcCtx.MyDB
	//var terms []string
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("type = '%s'", req.Type))
		db = db.Where("type = ?", req.Type)
	}
	//term := strings.Join(terms, " AND ")

	if err = db.Table("au_roles").Find(&items).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}
