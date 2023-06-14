package systemparam

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemParamUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemParamUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemParamUpdateLogic {
	return SystemParamUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemParamUpdateLogic) SystemParamUpdate(req *types.SystemParamUpdateRequest) error {
	for _, update := range req.List {
		l.svcCtx.MyDB.Table("bs_system_params").Where("name = ?", update.Name).
			Updates(map[string]interface{}{"value": update.Value})
	}
	return nil
}
