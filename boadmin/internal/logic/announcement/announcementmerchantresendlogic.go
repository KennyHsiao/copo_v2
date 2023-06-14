package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/skypeService"
	"com.copo/bo_service/boadmin/internal/service/telegramService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementMerchantResendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementMerchantResendLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementMerchantResendLogic {
	return AnnouncementMerchantResendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementMerchantResendLogic) AnnouncementMerchantResend(req *types.AnnouncementMerchantResendRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.AnnouncementId)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	announcementMerchant, err := model.NewAnnouncementMerchant(l.svcCtx.MyDB).GetAnnouncementMerchant(req.AnnouncementId, req.MerchantCode)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	messageId := announcementMerchant.MessageId

	// 只有失敗對象 才可以重送
	if announcementMerchant.Status != "3" {
		return errorz.New(response.STATUS_ERROR_FORBIDDEN_TO_SEND)
	}
	status := ""
	if announcement.CommunicationSoftware == "telegram" {

		if messageId != "" {
			_, err2 := telegramService.EditMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcementMerchant.MessageId, announcement.Content)
			if err2 != nil {
				return errorz.New(response.ANNOUNCEMENT_FAILED_TO_SEND)
			}
		} else {
			sendResp, err2 := telegramService.SendMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcement.Content)
			if err2 != nil {
				return errorz.New(response.ANNOUNCEMENT_FAILED_TO_SEND)
			}
			messageId = sendResp.Data.Msg
		}
		status = "2"
	} else if announcement.CommunicationSoftware == "skype" && messageId == "" {
		_, err3 := skypeService.SendMessage(l.ctx, l.svcCtx, strconv.FormatInt(announcementMerchant.ID, 10), announcementMerchant.ChatId, announcement.Content)
		if err3 != nil {
			return errorz.New(response.ANNOUNCEMENT_FAILED_TO_SEND)
		}
		status = "7" //處理中
	}

	if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
		Where("id = ?", announcementMerchant.ID).
		Updates(map[string]interface{}{
			"status":     status,
			"message_id": messageId,
		}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE)
	}

	// 判斷主表狀態
	err = model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(req.AnnouncementId)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
