package channeldata

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelDataQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataQueryLogic {
	return ChannelDataQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataQueryLogic) ChannelDataQuery(req types.ChannelDataQueryRequest) (resp *types.ChannelDataQueryResponse, err error) {
	var channel types.ChannelData
	channelPayTypeList := []types.ChannelPayType{}
	bankMapList := []types.BankCodeMap{}
	db := l.svcCtx.MyDB

	if req.ID > 0 {
		db = db.Where("id = ?", req.ID)
	}
	if len(req.Code) > 0 {
		db = db.Where("code = ?", req.Code)
	}
	err = db.Table("ch_channels").
		Preload("ChannelPayTypeList").
		Preload("Banks", "currency_code=(?)", req.CurrencyCode).
		Find(&channel).Error

	if len(channel.ChannelPayTypeList) > 0 {
		err = l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("channel_code=?", channel.Code).Find(&channelPayTypeList).Error
	}

	if len(channel.Banks) > 0 {
		err = l.svcCtx.MyDB.Table("ch_channel_banks").Where("channel_code", channel.Code).Find(&bankMapList).Error
	}

	channel.BankCodeMapList = bankMapList
	channel.ChannelPayTypeList = channelPayTypeList
	resp = &types.ChannelDataQueryResponse{
		channel,
	}

	return resp, err
}
