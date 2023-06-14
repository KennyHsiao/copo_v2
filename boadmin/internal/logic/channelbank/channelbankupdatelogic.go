package channelbank

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBankUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBankUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBankUpdateLogic {
	return ChannelBankUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBankUpdateLogic) ChannelBankUpdate(req types.ChannelBankUpdateRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
