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

type AnnouncementResendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementResendLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementResendLogic {
	return AnnouncementResendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementResendLogic) AnnouncementResend(req *types.AnnouncementResendRequest) error {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有失敗狀態才能重送
	if announcement.Status != "3" {
		return errorz.New(response.STATUS_ERROR_FORBIDDEN_TO_SEND)
	}
	announcementMerchants, err := model.NewAnnouncementMerchant(l.svcCtx.MyDB).FindByAnnouncementIdAndStatus(announcement.ID, "3")
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	for _, announcementMerchant := range announcementMerchants {
		status := ""
		// 只有失敗對象 才可以重送
		if announcementMerchant.Status != "3" {
			continue
		}
		messageId := announcementMerchant.MessageId
		if announcement.CommunicationSoftware == "telegram" {
			if messageId != "" {
				_, err2 := telegramService.EditMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcementMerchant.MessageId, announcement.Content)
				if err2 != nil {
					continue
				}
			} else {
				sendResp, err2 := telegramService.SendMessage(l.ctx, l.svcCtx, announcementMerchant.ChatId, announcement.Content)
				if err2 != nil {
					continue
				}
				messageId = sendResp.Data.Msg
			}
			status = "2"
		} else if announcement.CommunicationSoftware == "skype" && messageId == "" {
			_, err3 := skypeService.SendMessage(l.ctx, l.svcCtx, strconv.FormatInt(announcementMerchant.ID, 10), announcementMerchant.ChatId, announcement.Content)
			if err3 != nil {
				continue
			}
			status = "7" //處理中
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
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
