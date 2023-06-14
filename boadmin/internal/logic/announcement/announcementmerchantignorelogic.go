package announcement

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementMerchantIgnoreLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementMerchantIgnoreLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementMerchantIgnoreLogic {
	return AnnouncementMerchantIgnoreLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementMerchantIgnoreLogic) AnnouncementMerchantIgnore(req *types.AnnouncementMerchantIgnoreRequest) error {
	announcementMerchant, err := model.NewAnnouncementMerchant(l.svcCtx.MyDB).GetAnnouncementMerchant(req.AnnouncementId, req.MerchantCode)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 只有失敗對象 且 沒有 MessageId 才可以忽略
	if announcementMerchant.Status != "3" || announcementMerchant.MessageId != "" {
		return errorz.New(response.STATUS_ERROR_OPERATION_PROHIBITED)
	}

	err = l.svcCtx.MyDB.Table("an_announcement_merchants").
		Where("id = ?", announcementMerchant.ID).
		Updates(map[string]interface{}{
			"status": "6",
		}).Error
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 判斷主表狀態
	err = model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(req.AnnouncementId)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return err
}
