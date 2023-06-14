package channelbank

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBankQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBankQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBankQueryLogic {
	return ChannelBankQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBankQueryLogic) ChannelBankQuery(req types.ChannelBankQueryRequest) (resp *types.ChannelBankQueryResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
