package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementDeleteLogic {
	return AnnouncementDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementDeleteLogic) AnnouncementDelete(req *types.AnnouncementDeleteRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有草稿和回收狀態才能刪除
	if announcement.Status != "1" && announcement.Status != "4" {
		return errorz.New(response.STATUS_ERROR_FORBIDDEN_TO_DELETE)
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		if err = l.svcCtx.MyDB.Table("an_announcements").
			Delete(announcement).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err = l.svcCtx.MyDB.Table("an_announcement_channels").
			Where("announcement_id = ?", announcement.ID).
			Delete(&types.AnnouncementChannelX{}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
			Where("announcement_id = ?", announcement.ID).
			Delete(&types.AnnouncementMerchantX{}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err = l.svcCtx.MyDB.Table("an_announcement_params").
			Where("announcement_id = ?", announcement.ID).
			Delete(&types.AnnouncementParamX{}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return err
	})
}
