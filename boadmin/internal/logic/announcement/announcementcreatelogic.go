package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/skypeService"
	"com.copo/bo_service/boadmin/internal/service/telegramService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementCreateLogic {
	return AnnouncementCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementCreateLogic) AnnouncementCreate(req types.AnnouncementCreateRequest) error {

	var announcementMerchants []*types.AnnouncementMerchantX
	if req.Action == "send" {
		req.Status = "3"
	} else {
		//保存為草稿
		req.Status = "1"
	}

	announcement := &types.AnnouncementCreate{
		AnnouncementCreateRequest: req,
	}

	if err := l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		if err = l.svcCtx.MyDB.Table("an_announcements").Create(announcement).Error; err != nil {
			return
		}

		for _, merchant := range req.AnnouncementMerchants {
			merchant.AnnouncementId = announcement.ID
			merchantCreate := &types.AnnouncementMerchantX{
				AnnouncementMerchant: merchant,
			}
			merchantCreate.Status = "1"
			if err = l.svcCtx.MyDB.Table("an_announcement_merchants").Create(merchantCreate).Error; err != nil {
				return
			}
			announcementMerchants = append(announcementMerchants, merchantCreate)
		}

		for _, channel := range req.AnnouncementChannels {
			channel.AnnouncementId = announcement.ID
			channelCreate := &types.AnnouncementChannelX{
				AnnouncementChannel: channel,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_channels").Create(channelCreate).Error; err != nil {
				return
			}
		}

		for _, param := range req.AnnouncementParams {
			param.AnnouncementId = announcement.ID
			paramCreate := &types.AnnouncementParamX{
				AnnouncementParam: param,
			}
			if err = l.svcCtx.MyDB.Table("an_announcement_params").Create(paramCreate).Error; err != nil {
				return
			}
		}

		return
	}); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if req.Action == "send" {

		type MessageData struct {
			SendStatus string
			MessageId  string
		}
		merchantMap := map[string]MessageData{}

		for _, announcementMerchant := range announcementMerchants {
			// 取得商戶
			var merchant *types.Merchant
			var err2 error
			if merchant, err2 = model.NewMerchant(l.svcCtx.MyDB).GetMerchantByCode(announcementMerchant.MerchantCode); err2 != nil {
				logx.WithContext(l.ctx).Errorf("取得商戶資料錯誤:%s,%s", announcementMerchant.MerchantCode, err2.Error())
				announcementMerchant.Status = "3"
				l.svcCtx.MyDB.Table("an_announcement_merchants").Updates(announcementMerchant)
				continue
			}

			chatId := merchant.Contact.GroupID
			// 對象的CommunicationSoftware 與公告不匹配  或 沒chatId 則忽略
			if merchant.Contact.CommunicationSoftware != req.CommunicationSoftware || chatId == "" {
				logx.WithContext(l.ctx).Errorf("不支援發送通知:MerchantCode:%s,CommunicationSoftware:%s,GroupID:%s",
					announcementMerchant.MerchantCode, merchant.Contact.CommunicationSoftware, chatId)
				announcementMerchant.Status = "6"
				l.svcCtx.MyDB.Table("an_announcement_merchants").Updates(announcementMerchant)
				continue
			}

			//重複chatId只發送一次; map裡是已發送過,資料要同步
			if messageData, ok := merchantMap[chatId]; ok {
				announcementMerchant.Status = "6"
				announcementMerchant.ChatId = chatId
				announcementMerchant.MessageId = messageData.MessageId
				l.svcCtx.MyDB.Table("an_announcement_merchants").Updates(announcementMerchant)
				continue
			}

			// telegram
			if req.CommunicationSoftware == "telegram" {
				sendResp, err3 := telegramService.SendMessage(l.ctx, l.svcCtx, chatId, req.Content)
				if err3 != nil {
					announcementMerchant.Status = "3" //發送失敗
				} else {
					announcementMerchant.Status = "2" //發送成功
				}
				announcementMerchant.ChatId = chatId
				announcementMerchant.MessageId = sendResp.Data.Msg
			}
			// skype
			if req.CommunicationSoftware == "skype" {
				_, err3 := skypeService.SendMessage(l.ctx, l.svcCtx, strconv.FormatInt(announcementMerchant.ID, 10), chatId, req.Content)
				if err3 != nil {
					announcementMerchant.Status = "3" //發送失敗
				} else {
					announcementMerchant.Status = "7" //處理中
				}
				announcementMerchant.ChatId = chatId
			}

			//變更發送狀態
			l.svcCtx.MyDB.Table("an_announcement_merchants").Updates(announcementMerchant)

			//保存此chatId的紀錄
			merchantMap[chatId] = MessageData{
				SendStatus: announcementMerchant.Status,
				MessageId:  announcementMerchant.MessageId,
			}
		}
		// 判斷主表狀態
		err := model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(announcement.ID)
		if err != nil {
			logx.WithContext(l.ctx).Errorf("主表編輯狀態失敗 ,err:%s", err.Error())
		}
	}

	return nil
}
