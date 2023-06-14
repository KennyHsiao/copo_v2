package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	"context"
	"github.com/zeromicro/go-zero/core/logx"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
)

type ChannelPayTypeQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeQueryAllLogic {
	return ChannelPayTypeQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeQueryAllLogic) ChannelPayTypeQueryAll(req types.ChannelPayTypeQueryAllRequestX) (resp *types.ChannelPayTypeQueryAllResponse, err error) {
	return model.NewChannelPayType(l.svcCtx.MyDB).ChannelPayTypeQueryAll(req)
}
