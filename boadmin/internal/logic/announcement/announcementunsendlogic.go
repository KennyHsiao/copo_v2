package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/telegramService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementUnsendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementUnsendLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementUnsendLogic {
	return AnnouncementUnsendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementUnsendLogic) AnnouncementUnsend(req *types.AnnouncementUnsendRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有 telegram 才能回收
	if announcement.CommunicationSoftware != "telegram" {
		return errorz.New(response.STATU_ERROR_PROHIBIT_RECYCLING)
	}
	// 只有 失敗 成功 回收失敗 狀態才能回收
	if announcement.Status != "2" && announcement.Status != "3" && announcement.Status != "5" {
		return errorz.New(response.STATU_ERROR_PROHIBIT_RECYCLING)
	}

	announcementMerchants, err := model.NewAnnouncementMerchant(l.svcCtx.MyDB).FindByAnnouncementId(announcement.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	for _, announcementMerchant := range announcementMerchants {
		if announcementMerchant.Status == "4" || announcementMerchant.Status == "6" {
			//已回收或忽略 不做處理
			continue
		}
		if announcementMerchant.Status == "1" || announcementMerchant.MessageId == "" {
			// 草稿 或 沒有MessageId 直接改成忽略
			if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
				Where("id = ?", announcementMerchant.ID).
				Updates(map[string]interface{}{
					"status": "6",
				}).Error; err != nil {
				logx.WithContext(l.ctx).Errorf("狀態改忽略失敗 merchantCode:%s,err:%s", announcementMerchant.MerchantCode, err.Error())
			}
			continue
		}
		status := "4" //回收
		_, err = telegramService.DeleteMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcementMerchant.MessageId)
		if err != nil {
			status = "5" //回收失敗
			logx.WithContext(l.ctx).Errorf("收回訊息失敗 merchantCode:%s,err:%s", announcementMerchant.MerchantCode, err.Error())
		}
		if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
			Where("id = ?", announcementMerchant.ID).
			Updates(map[string]interface{}{
				"status": status,
			}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("狀態變更失敗 merchantCode:%s,err:%s", announcementMerchant.MerchantCode, err.Error())
		}
		// 判斷主表狀態
		err = model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(announcement.ID)
		if err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	return nil
}
