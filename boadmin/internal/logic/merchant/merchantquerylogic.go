package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryLogic {
	return MerchantQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryLogic) MerchantQuery(req types.MerchantQueryRequest) (resp *types.MerchantQueryResponse, err error) {

	ux := model.NewMerchant(l.svcCtx.MyDB)
	var merchant *types.Merchant

	if merchant, err = ux.GetMerchantByCode(req.Code); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	merchant.RegisteredAtS = utils.ParseIntTime(merchant.RegisteredAt)
	//merchant.CreatedAt = common.ParseTime(merchant.CreatedAt)

	if len(merchant.Users) > 0 {
		for i, _ := range merchant.Users {
			merchant.Users[i].RegisteredAtS = utils.ParseIntTime(merchant.Users[i].RegisteredAt)
		}
	}

	resp = &types.MerchantQueryResponse{
		Merchant: *merchant,
	}
	return
}
