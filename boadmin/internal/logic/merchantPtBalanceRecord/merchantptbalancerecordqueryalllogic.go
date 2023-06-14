package merchantPtBalanceRecord

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceRecordQueryAllLogic {
	return MerchantPtBalanceRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceRecordQueryAllLogic) MerchantPtBalanceRecordQueryAll(req *types.MerchantPtBalanceRecordQueryAllRequest) (resp *types.MerchantPtBalanceRecordQueryAllResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
