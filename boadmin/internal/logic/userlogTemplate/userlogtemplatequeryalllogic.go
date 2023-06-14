package userlogTemplate

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogTemplateQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogTemplateQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogTemplateQueryAllLogic {
	return UserLogTemplateQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogTemplateQueryAllLogic) UserLogTemplateQueryAll(req types.UserLogTemplateQueryAllRequestX) (resp *types.UserLogTemplateQueryAllResponseX, err error) {
	var temps []types.UserLogTemplateX
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	db2 := l.svcCtx.MyDB
	if len(req.ApiUnit) > 0 {
		//terms = append(terms, fmt.Sprintf("api_unit like '%%%s%%'", req.ApiUnit))
		db = db.Where("api_unit like ?", "%"+req.ApiUnit+"%")
		db2 = db2.Where("api_unit like ?", "%"+req.ApiUnit+"%")
	}
	if len(req.UserType) > 0 {
		//terms = append(terms, fmt.Sprintf("user_type = '%s'", req.UserType))
		db = db.Where("user_type = ?", req.UserType)
		db2 = db2.Where("user_type = ?", req.UserType)
	}
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("`type` = '%s'", req.Type))
		db = db.Where("type = ?", req.Type)
		db2 = db2.Where("user_type = ?", req.UserType)
	}
	if len(req.MsgTemplate) > 0 {
		//terms = append(terms, fmt.Sprintf("`msg_template` like '%%%s%%'", req.MsgTemplate))
		db = db.Where("msg_template like ?", "%"+req.MsgTemplate+"%")
		db2 = db2.Where("msg_template like ?", "%"+req.MsgTemplate+"%")
	}
	if len(req.Path) > 0 {
		//terms = append(terms, fmt.Sprintf("`path` = '%s'", req.Path))
		db = db.Where("path = ?", req.Path)
		db2 = db2.Where("path = ?", req.Path)
	}
	if len(req.ApiName) > 0 {
		//terms = append(terms, fmt.Sprintf("`api_name` like '%%%s%%'", req.ApiName))
		db = db.Where("api_name like ?", "%"+req.ApiName+"%")
		db2 = db2.Where("api_name like ?", "%"+req.ApiName+"%")
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("`updated_at` >= '%s'", req.StartAt))
		db = db.Where("updated_at >= ?", req.StartAt)
		db2 = db2.Where("updated_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("`updated_at` < '%s'", endAt))
		db = db.Where("updated_at < ?", endAt)
		db2 = db2.Where("updated_at >= ?", req.StartAt)
	}

	//term := strings.Join(terms, " AND ")
	db2.Table("au_user_log_template").Count(&count)
	err = db.Table("au_user_log_template").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&temps).Error

	resp = &types.UserLogTemplateQueryAllResponseX{
		List:     temps,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
