package announcement

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementQueryAllLogic {
	return AnnouncementQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementQueryAllLogic) AnnouncementQueryAll(req types.AnnouncementQueryAllRequestX) (resp *types.AnnouncementQueryAllResponseX, err error) {
	var announcements []types.AnnouncementX
	var count int64

	db := l.svcCtx.MyDB.Table("an_announcements")

	if len(req.TemplateCode) > 0 {
		db = db.Where("an_announcements.template_code in ? ", req.TemplateCode)
	}
	if len(req.Title) > 0 {
		db = db.Where("an_announcements.title like ?", "%"+req.Title+"%")
	}
	if len(req.Status) > 0 {
		db = db.Where("an_announcements.status in ? ", req.Status)
	}
	if len(req.Content) > 0 {
		db = db.Where("an_announcements.content like ?", "%"+req.Content+"%")
	}
	if len(req.StartAt) > 0 {
		db = db.Where("an_announcements.updated_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		db = db.Where("an_announcements.updated_at < ?", endAt)
	}
	if len(req.MerchantCode) > 0 {
		db.Joins("join an_announcement_merchants am on am.announcement_id = an_announcements.id ").
			Group("an_announcements.id").Where("am.merchant_code like ?", "%"+req.MerchantCode+"%")
	}

	err = db.Count(&count).Error

	err = db.Select("an_announcements.*").
		Preload("AnnouncementChannels").
		Preload("AnnouncementMerchants.Merchant").
		Preload("AnnouncementParams").
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&announcements).Error

	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.AnnouncementQueryAllResponseX{
		List:     announcements,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return

}
