package merchant

import (
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateLogic {
	return MerchantUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateLogic) MerchantUpdate(req types.MerchantUpdateRequest) error {
	merchant := &types.MerchantUpdate{
		MerchantUpdateRequest: req,
	}
	return l.svcCtx.MyDB.Table("mc_merchants").Session(&gorm.Session{FullSaveAssociations: true}).Where("code = ?", req.Code).Updates(merchant).Error
}
