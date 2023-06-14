package channeldata

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelDataGetDropDownListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataGetDropDownListLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataGetDropDownListLogic {
	return ChannelDataGetDropDownListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataGetDropDownListLogic) ChannelDataGetDropDownList(req types.ChannelDataGetDropDownListRequest) (items []types.ChannelDataDropDownItem, err error) {

	db := l.svcCtx.MyDB.Table("ch_channels").Order("code")

	if len(req.Code) > 0 {
		db.Where("code like ", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		db.Where("name like ", "%"+req.Name+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db.Where("currency_code = ?", req.CurrencyCode)
	}
	if len(req.Status) > 0 {
		db.Where("status IN ?", req.Status)
	}

	err = db.Find(&items).Error

	return items, err
}
