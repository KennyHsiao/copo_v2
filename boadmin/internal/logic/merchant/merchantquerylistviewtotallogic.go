package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryListViewTotalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryListViewTotalLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryListViewTotalLogic {
	return MerchantQueryListViewTotalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryListViewTotalLogic) MerchantQueryListViewTotal(req *types.MerchantQueryListViewRequestX) (resp *types.MerchantQueryListViewTotalResponse, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}
	return model.NewMerchantListView(l.svcCtx.MyDB).QueryListViewTotal(*req)
}
