package systemparam

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemParamQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemParamQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemParamQueryAllLogic {
	return SystemParamQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemParamQueryAllLogic) SystemParamQueryAll(req *types.SystemParamQueryAllRequest) (resp *types.SystemParamQueryAllResponse, err error) {
	var systemParam []types.SystemParam

	err = l.svcCtx.MyDB.Table("bs_system_params").Find(&systemParam).Error

	resp = &types.SystemParamQueryAllResponse{
		List: systemParam,
	}
	return
}
