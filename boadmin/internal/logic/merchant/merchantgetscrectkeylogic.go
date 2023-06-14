package merchant

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantGetScrectKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantGetScrectKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantGetScrectKeyLogic {
	return MerchantGetScrectKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantGetScrectKeyLogic) MerchantGetScrectKey(req types.MerchantGetScrectKeyRequest) (resp *types.MerchantGetScrectKeyResponse, err error) {
	if err = l.svcCtx.MyDB.Table("mc_merchants").
		Select("screct_key").
		Where("code = ?", req.Code).
		Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}
