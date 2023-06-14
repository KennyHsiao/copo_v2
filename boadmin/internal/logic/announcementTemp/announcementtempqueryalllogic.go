package announcementTemp

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementTempQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementTempQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementTempQueryAllLogic {
	return AnnouncementTempQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementTempQueryAllLogic) AnnouncementTempQueryAll(req types.AnnouncementTempQueryAllRequestX) (resp *types.AnnouncementTempQueryAllResponse, err error) {
	var announcementTemps []types.AnnouncementTemp
	var count int64
	db := l.svcCtx.MyDB

	if err = db.Table("an_templates").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("an_templates").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&announcementTemps).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.AnnouncementTempQueryAllResponse{
		List:     announcementTemps,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
