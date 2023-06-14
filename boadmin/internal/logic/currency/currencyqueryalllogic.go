package currency

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyQueryAllLogic {
	return CurrencyQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyQueryAllLogic) CurrencyQueryAll(req types.CurrencyQueryAllRequestX) (resp *types.CurrencyQueryAllResponse, err error) {
	return model.NewCurrency(l.svcCtx.MyDB).CurrencyQueryAll(req)
}
