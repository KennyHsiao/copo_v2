package channelbalance

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBalanceQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBalanceQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBalanceQueryLogic {
	return ChannelBalanceQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBalanceQueryLogic) ChannelBalanceQuery(req *types.ChannelBalanceQueryRequest) (resp *types.ChannelBalanceQueryResponse, err error) {
	db := l.svcCtx.MyDB

	if len(req.Code) > 0 {
		db = db.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		db = db.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.Status) > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if len(req.IsProxy) > 0 {
		db = db.Where("is_proxy = ?", req.IsProxy)
	} else {
		db = db.Where("(is_proxy = '1' OR is_proxy = '0')")
	}
	db = db.Where("currency_code = ?", req.CurrencyCode)
	db = db.Where("status != '0'")

	selectX := "SUM(withdraw_balance) as xf_balance," +
		"SUM(proxypay_balance) as df_balance," +
		"SUM(withdraw_balance+proxypay_balance) as total_balance"

	if err = db.Table("ch_channels").
		Select(selectX).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
