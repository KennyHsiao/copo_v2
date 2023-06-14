package currency

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyDeleteLogic {
	return CurrencyDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyDeleteLogic) CurrencyDelete(req types.CurrencyDeleteRequest) (err error) {

	if err := l.svcCtx.MyDB.Table("bs_currencies").Delete(&req).Error; err != nil {
		return errorz.New(response.DELETE_DATABASE_FAILURE, err.Error())
	}

	return err
}
