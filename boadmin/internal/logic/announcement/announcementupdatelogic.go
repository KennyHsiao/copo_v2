package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/telegramService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementUpdateLogic {
	return AnnouncementUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementUpdateLogic) AnnouncementUpdate(req *types.AnnouncementUpdateRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有 telegram 才能編輯重發
	if announcement.CommunicationSoftware != "telegram" {
		return errorz.New(response.STATUS_ERROR_OPERATION_PROHIBITED)
	}
	// 只有 失敗 成功 才能編輯重發
	if announcement.Status != "2" && announcement.Status != "3" {
		return errorz.New(response.STATUS_ERROR_OPERATION_PROHIBITED)
	}

	announcementUpdate := &types.AnnouncementUpdate{
		AnnouncementUpdateRequest: *req,
	}

	if err = l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		// 刪除渠道
		if err = l.svcCtx.MyDB.Table("an_announcement_channels").
			Where("announcement_id = ?", announcement.ID).
			Delete(&types.AnnouncementChannelX{}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		// 刪除參數
		if err = l.svcCtx.MyDB.Table("an_announcement_params").
			Where("announcement_id = ?", announcement.ID).
			Delete(&types.AnnouncementParamX{}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 編輯主表
		announcementUpdate.Status = "3"
		if err = l.svcCtx.MyDB.Table("an_announcements").
			Omit("AnnouncementChannels", "AnnouncementParams", "AnnouncementMerchants").
			Updates(announcementUpdate).Error; err != nil {
			return
		}
		// 建立渠道
		for _, channel := range req.AnnouncementChannels {
			channel.AnnouncementId = announcement.ID
			channelCreate := &types.AnnouncementChannelX{
				AnnouncementChannel: channel,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_channels").Create(channelCreate).Error; err != nil {
				return
			}
		}
		// 建立參數
		for _, param := range req.AnnouncementParams {
			param.AnnouncementId = announcement.ID
			paramCreate := &types.AnnouncementParamX{
				AnnouncementParam: param,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_params").Create(paramCreate).Error; err != nil {
				return
			}
		}

		return err
	}); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for _, announcementMerchant := range announcement.AnnouncementMerchants {
		messageId := announcementMerchant.MessageId
		var err2 error
		status := "2"
		var sendResp types.TelegramSendResponse
		if announcementMerchant.ChatId == "" || announcementMerchant.Status == "6" {
			// ChatId為空 或 狀態為忽略 不執行
			continue
		}
		if messageId != "" {
			//曾經發過訊息則編輯
			_, err2 = telegramService.EditMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcementMerchant.MessageId, req.Content)
		} else {
			//曾經沒成功發過訊息 則再發一則
			sendResp, err2 = telegramService.SendMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, req.Content)
			messageId = sendResp.Data.Msg
		}
		if err2 != nil {
			logx.WithContext(l.ctx).Errorf("發送訊息失敗 merchantCode:%s,err:%s", announcementMerchant.MerchantCode, err2.Error())
			status = "3"
		}
		if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
			Where("id = ?", announcementMerchant.ID).
			Updates(map[string]interface{}{
				"status":     status,
				"message_id": messageId,
			}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("發送訊息後編輯狀態失敗 merchantCode:%s,err:%s", announcementMerchant.MerchantCode, err.Error())
		}

	}
	// 判斷主表狀態
	err = model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(announcement.ID)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("主表編輯狀態失敗,err:%s", err.Error())
	}

	return nil
}
