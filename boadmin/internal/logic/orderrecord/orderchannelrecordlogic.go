package orderrecord

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderChannelRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderChannelRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderChannelRecordLogic {
	return OrderChannelRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderChannelRecordLogic) OrderChannelRecord(req types.OrderChannelRecordReqeust) (resp *types.OrderChannelRecordResponse, err error) {
	var orderChannels []types.OrderChannels

	if err = l.svcCtx.MyDB.Table("tx_order_channels").Where("`order_no` = ?", &req.OrderNo).Find(&orderChannels).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.OrderChannelRecordResponse{
		List: orderChannels,
	}
	return
}
