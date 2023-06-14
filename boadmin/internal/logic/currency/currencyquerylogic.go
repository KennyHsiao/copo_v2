package currency

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyQueryLogic {
	return CurrencyQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyQueryLogic) CurrencyQuery(req types.CurrencyQueryRequest) (resp *types.CurrencyQueryResponse, err error) {

	if err = l.svcCtx.MyDB.Table("bs_currencies").Take(&resp, req.ID).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//err = l.svcCtx.MyDB.Table("bs_currencies").Take(&resp, req.ID).Error;

	return
}
