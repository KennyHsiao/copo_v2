package channelbank

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBankDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBankDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBankDeleteLogic {
	return ChannelBankDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBankDeleteLogic) ChannelBankDelete(req types.ChannelBankDeleteRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
