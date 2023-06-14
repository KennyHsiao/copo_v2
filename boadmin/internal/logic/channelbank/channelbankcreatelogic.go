package channelbank

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBankCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBankCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBankCreateLogic {
	return ChannelBankCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBankCreateLogic) ChannelBankCreate(req types.ChannelBankCreateRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
