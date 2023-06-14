package timezonex

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimezonexQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimezonexQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) TimezonexQueryAllLogic {
	return TimezonexQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimezonexQueryAllLogic) TimezonexQueryAll(req types.TimezonexQueryAllRequest) (resp *types.TimezonexQueryAllResponse, err error) {
	var timezones []types.Timezonex
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	db2 := l.svcCtx.MyDB
	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name like '%%%s%%'", req.Name))
		db = db.Where("name like ?", "%"+req.Name+"%")
		db2 = db2.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.Code) > 0 {
		//terms = append(terms, fmt.Sprintf("code like '%%%s%%'", req.Code))
		db = db.Where("code like ?", "%"+req.Code+"%")
		db2 = db2.Where("code like ?", "%"+req.Code+"%")
	}
	//term := strings.Join(terms, "AND")
	db2.Table("bs_timezone").Count(&count)
	err = db.Table("bs_timezone").Scopes(gormx.Paginate(req)).Find(&timezones).Error

	resp = &types.TimezonexQueryAllResponse{
		List:     timezones,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
