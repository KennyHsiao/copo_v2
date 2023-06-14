package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantGetSubAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantGetSubAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantGetSubAgentsLogic {
	return MerchantGetSubAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantGetSubAgentsLogic) MerchantGetSubAgents() (resp []types.Merchant, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode == "" {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	return model.NewMerchant(l.svcCtx.MyDB).GetSubAgents(jwtMerchantCode)
}
