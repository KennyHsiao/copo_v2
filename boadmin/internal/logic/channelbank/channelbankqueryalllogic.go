package channelbank

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBankQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBankQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBankQueryAllLogic {
	return ChannelBankQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBankQueryAllLogic) ChannelBankQueryAll(req types.ChannelBankQueryAllRequest) (resp *types.ChannelBankQueryAllResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
