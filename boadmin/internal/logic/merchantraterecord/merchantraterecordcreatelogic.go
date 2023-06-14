package merchantraterecord

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantRateRecordCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantRateRecordCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantRateRecordCreateLogic {
	return MerchantRateRecordCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantRateRecordCreateLogic) MerchantRateRecordCreate(req *types.MerchantRateRecordCreateRequest) error {
	merchantRateRecordCreate := &types.MerchantRateRecordCreate{
		MerchantRateRecordCreateRequest: *req,
	}

	return l.svcCtx.MyDB.Table("mc_merchant_rate_record").Create(merchantRateRecordCreate).Error
}
