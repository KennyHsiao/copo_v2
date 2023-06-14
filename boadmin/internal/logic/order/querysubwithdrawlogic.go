package order

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QuerySubWithdrawLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQuerySubWithdrawLogic(ctx context.Context, svcCtx *svc.ServiceContext) QuerySubWithdrawLogic {
	return QuerySubWithdrawLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QuerySubWithdrawLogic) QuerySubWithdraw(req *types.QuerySubWithdrawRequest) (resp *types.QuerySubWithdrawResponse, err error) {
	var list []types.TxOrderChannels

	//var txOrder types.OrderX

	//if errTx := l.svcCtx.MyDB.Table("tx_orders tx").
	//	Where("order_no = ?",req.OrderNo).
	//	Take(&txOrder).Error;errTx != nil{
	//	return nil, errorz.New(response.DATABASE_FAILURE)
	//}

	if err := l.svcCtx.MyDB.Table("tx_order_channels toc").
		Joins("left join ch_channels cc ON toc.channel_code = cc.code").
		Where("order_no like ?", "%"+req.OrderNo+"%").
		Select("cc.name AS channel_name,toc.*").
		Find(&list).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.DATA_NOT_FOUND)
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}
	}

	for i, orderChannel := range list {
		var channelPayType types.ChannelPayType
		if errTx := l.svcCtx.MyDB.Table("ch_channel_pay_types").
			Where("code = ?", orderChannel.ChannelCode+"DF").
			Take(&channelPayType).Error; errTx != nil {
			return nil, errorz.New(response.DATABASE_FAILURE)
		}
		list[i].IsRate = channelPayType.IsRate
	}

	return &types.QuerySubWithdrawResponse{List: list}, nil
}
