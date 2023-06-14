package merchant

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendNotifyChannelsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendNotifyChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) SendNotifyChannelsLogic {
	return SendNotifyChannelsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendNotifyChannelsLogic) SendNotifyChannels(req *types.MerchantChannelNotifyRequest) error {

	err := merchantsService.ChannelChangeNotify(l.svcCtx.MyDB, l.ctx, l.svcCtx, req.CurrencyCode)

	if err != nil {
		return err
	}

	return nil
}
