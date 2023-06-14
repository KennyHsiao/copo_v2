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

type AnnouncementTempParamQueryAllAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementTempParamQueryAllAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementTempParamQueryAllAllLogic {
	return AnnouncementTempParamQueryAllAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementTempParamQueryAllAllLogic) AnnouncementTempParamQueryAllAll(req types.AnnouncementTempParamQueryAllRequestX) (resp *types.AnnouncementTempParamQueryAllResponse, err error) {
	var announcementTempParams []types.AnnouncementTempParam
	var count int64
	db := l.svcCtx.MyDB

	if len(req.TemplateCode) > 0 {
		db = db.Where("`template_code` = ?", req.TemplateCode)
	}
	if len(req.Language) > 0 {
		db = db.Where("`language` = ?", req.Language)
	}

	if err = db.Table("an_template_params").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("an_template_params").
		Scopes(gormx.Sort(req.Orders)).Find(&announcementTempParams).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.AnnouncementTempParamQueryAllResponse{
		List: announcementTempParams,
	}

	return
}
