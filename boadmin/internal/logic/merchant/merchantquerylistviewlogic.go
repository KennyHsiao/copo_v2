package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryListViewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryListViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryListViewLogic {
	return MerchantQueryListViewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryListViewLogic) MerchantQueryListView(req types.MerchantQueryListViewRequestX) (resp *types.MerchantQueryListViewResponse, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}
	return model.NewMerchantListView(l.svcCtx.MyDB).QueryListView(req)
}
