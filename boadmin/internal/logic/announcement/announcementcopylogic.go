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

type AnnouncementCopyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementCopyLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementCopyLogic {
	return AnnouncementCopyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementCopyLogic) AnnouncementCopy(req *types.AnnouncementCopyRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	err = l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		// 複製主表
		announcement.ID = 0
		announcement.Status = "1"
		announcementX := &types.AnnouncementX{
			Announcement: *announcement,
		}
		if err = l.svcCtx.MyDB.Table("an_announcements").
			Omit("AnnouncementChannels", "AnnouncementParams", "AnnouncementMerchants").
			Create(announcementX).Error; err != nil {
			return
		}

		// 複製對象
		for _, merchant := range announcement.AnnouncementMerchants {
			merchant.ID = 0
			merchant.AnnouncementId = announcementX.ID
			merchant.Status = "1"
			merchant.ChatId = ""
			merchant.MessageId = ""
			merchantCreate := &types.AnnouncementMerchantX{
				AnnouncementMerchant: merchant,
			}
			merchantCreate.Status = "1"
			if err = l.svcCtx.MyDB.Table("an_announcement_merchants").Create(merchantCreate).Error; err != nil {
				return
			}
		}

		for _, channel := range announcement.AnnouncementChannels {
			channel.ID = 0
			channel.AnnouncementId = announcementX.ID
			channelCreate := &types.AnnouncementChannelX{
				AnnouncementChannel: channel,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_channels").Create(channelCreate).Error; err != nil {
				return
			}
		}

		for _, param := range announcement.AnnouncementParams {
			param.ID = 0
			param.AnnouncementId = announcementX.ID
			paramCreate := &types.AnnouncementParamX{
				AnnouncementParam: param,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_params").Create(paramCreate).Error; err != nil {
				return
			}
		}
		return
	})

	return err
}
