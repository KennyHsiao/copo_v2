package channelbalance

import (
	channelBalanceBalance "com.copo/bo_service/boadmin/internal/service/channelBalanceService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBalanceUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBalanceUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBalanceUpdateLogic {
	return ChannelBalanceUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBalanceUpdateLogic) ChannelBalanceUpdate(req *types.ChannelBalanceUpdateRequest) (err error) {
	if len(req.ChannelCodeList) <= 0 {
		return errorz.New(response.SETTING_CHANNEL_BALANCE_NULL)
	}

	err = channelBalanceBalance.UpdateChannelBalance(l.ctx, l.svcCtx, req)
	return
}
