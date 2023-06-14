package merchant

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateParamsRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateParamsRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateParamsRequestLogic {
	return MerchantUpdateParamsRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateParamsRequestLogic) MerchantUpdateParamsRequest(req *types.MerchantUpdateParamsRequest) error {

	return l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", req.Code).
		Updates(map[string]interface{}{
			"unpaid_notify_interval": req.UnpaidNotifyInterval,
			"unpaid_notify_num":      req.UnpaidNotifyNum},
		).Error
}
