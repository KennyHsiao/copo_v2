package merchantbindbank

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBindBankQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBindBankQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBindBankQueryLogic {
	return MerchantBindBankQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBindBankQueryLogic) MerchantBindBankQuery(req types.MerchantBindBankQueryRequest) (resp *types.MerchantBindBankQueryResponse, err error) {

	if err = l.svcCtx.MyDB.Table("mc_merchant_bind_bank").Take(&resp, req.ID).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
