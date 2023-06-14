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

type MerchantGetDescendantAgentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantGetDescendantAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantGetDescendantAgentsLogic {
	return MerchantGetDescendantAgentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantGetDescendantAgentsLogic) MerchantGetDescendantAgents(req *types.GetDescendantAgentsRequest) (resp []types.Merchant, err error) {

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode == "" {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}
	var merchant *types.Merchant
	ux := model.NewMerchant(l.svcCtx.MyDB)

	if merchant, err = ux.GetMerchantByCode(jwtMerchantCode); err != nil {
		return nil, errorz.New(response.MERCHANT_AGENT_NOT_FOUND)
	}

	return ux.GetDescendantAgents(merchant.AgentLayerCode, req.IsIncludeItself)
}
