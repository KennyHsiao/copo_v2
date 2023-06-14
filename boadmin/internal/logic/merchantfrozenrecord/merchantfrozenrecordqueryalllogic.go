package merchantfrozenrecord

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantFrozenRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantFrozenRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantFrozenRecordQueryAllLogic {
	return MerchantFrozenRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantFrozenRecordQueryAllLogic) MerchantFrozenRecordQueryAll(req types.MerchantFrozenRecordQueryAllRequest) (resp *types.MerchantFrozenRecordQueryAllResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
