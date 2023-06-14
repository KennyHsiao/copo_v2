package systemrate

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SystemChannelBalanceLimitQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemChannelBalanceLimitQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemChannelBalanceLimitQueryLogic {
	return SystemChannelBalanceLimitQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemChannelBalanceLimitQueryLogic) SystemChannelBalanceLimitQuery() (resp string, err error) {
	var systemParams types.SystemParams
	if err = l.svcCtx.MyDB.Table("bs_system_params").
		Where("name = 'channelBalanceLimit'").Find(&systemParams).Error; err != nil {
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = systemParams.Value
	return
}
