package merchant

import (
	"com.copo/bo_service/common/random"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantResetScrectKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantResetScrectKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantResetScrectKeyLogic {
	return MerchantResetScrectKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantResetScrectKeyLogic) MerchantResetScrectKey(req types.MerchantResetScrectKeyRequest) error {
	err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.Code).
		Update("screct_key", random.GetRandomString(32, random.ALL, random.MIX)).Error

	return err
}
