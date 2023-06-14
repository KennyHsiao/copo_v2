package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryAllLogic {
	return MerchantQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryAllLogic) MerchantQueryAll(req types.MerchantQueryAllRequestX) (resp *types.MerchantQueryAllResponse, err error) {
	var merchants []types.Merchant
	var count int64

	ux := model.NewMerchant(l.svcCtx.MyDB)
	if merchants, count, err = ux.QueryMerchants(req); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantQueryAllResponse{
		List:     merchants,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
