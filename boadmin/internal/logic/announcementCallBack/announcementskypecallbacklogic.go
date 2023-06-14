package announcementCallBack

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AnnouncementSkypeCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementSkypeCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementSkypeCallBackLogic {
	return AnnouncementSkypeCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementSkypeCallBackLogic) AnnouncementSkypeCallBack(req *types.SkypeCallBackRequest) (err error) {
	var announcementMerchant types.AnnouncementMerchant
	if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
		Take(&announcementMerchant, req.ID).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("get an_announcement_merchants error id:%s,err:%s", req.ID, err.Error())
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = l.svcCtx.MyDB.Table("an_announcement_merchants").
		Where("id = ?", req.ID).
		Updates(map[string]interface{}{
			"status":     "2",
			"message_id": req.MessageId,
		}).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("發送訊息後編輯狀態失敗 id:%s,err:%s", req.ID, err.Error())
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 判斷主表狀態
	err = model.NewAnnouncement(l.svcCtx.MyDB).AutoChangeStatus(announcementMerchant.AnnouncementId)
	if err != nil {
		logx.WithContext(l.ctx).Errorf("主表編輯狀態失敗,err:%s", err.Error())
	}

	return
}
