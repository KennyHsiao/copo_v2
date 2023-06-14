package channelraterecord

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelRateRecordCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRateRecordCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRateRecordCreateLogic {
	return ChannelRateRecordCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRateRecordCreateLogic) ChannelRateRecordCreate(req *types.ChannelRateRecordCreateRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
