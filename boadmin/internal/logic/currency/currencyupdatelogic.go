package currency

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyUpdateLogic {
	return CurrencyUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyUpdateLogic) CurrencyUpdate(req types.CurrencyUpdateRequest) (err error) {
	currency := &types.CurrencyUpdate{
		CurrencyUpdateRequest: req,
	}
	if err = l.svcCtx.MyDB.Table("bs_currencies").Updates(currency).Error; err != nil {
		return errorz.New(response.UPDATE_DATABASE_FAILURE, err.Error())
	}

	return
}
