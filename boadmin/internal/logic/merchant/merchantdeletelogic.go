package merchant

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantDeleteLogic {
	return MerchantDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantDeleteLogic) MerchantDelete(req types.MerchantDeleteRequest) error {
	return l.svcCtx.MyDB.Table("mc_merchants").Delete(&req).Error
}
