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

type AnnouncementQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAnnouncementQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) AnnouncementQueryLogic {
	return AnnouncementQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AnnouncementQueryLogic) AnnouncementQuery(req *types.AnnouncementQueryRequest) (resp *types.AnnouncementQueryResponse, err error) {
	announcement, err := model.NewAnnouncement(l.svcCtx.MyDB).GetAnnouncement(req.ID)
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	resp = &types.AnnouncementQueryResponse{
		Announcement: *announcement,
	}
	return
}
