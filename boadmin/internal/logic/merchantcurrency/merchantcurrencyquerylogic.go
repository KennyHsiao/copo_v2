package merchantcurrency

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantCurrencyQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantCurrencyQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantCurrencyQueryLogic {
	return MerchantCurrencyQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantCurrencyQueryLogic) MerchantCurrencyQuery(req types.MerchantCurrencyQueryRequest) (resp *types.MerchantCurrencyQueryResponse, err error) {
	if err = l.svcCtx.MyDB.Table("mc_merchant_currencies").Take(&resp, req.ID).Error; err != nil {
		return resp, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}
