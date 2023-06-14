package channeldata

import (
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelDataDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataDeleteLogic {
	return ChannelDataDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataDeleteLogic) ChannelDataDelete(req types.ChannelDataDeleteRequest) error {
	var channel types.ChannelData
	return l.svcCtx.MyDB.Table("ch_channels").Where("code = ?", req.Code).Delete(&channel).Error
}
