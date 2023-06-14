package systemrate

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemChannelBalanceLimitUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemChannelBalanceLimitUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemChannelBalanceLimitUpdateLogic {
	return SystemChannelBalanceLimitUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemChannelBalanceLimitUpdateLogic) SystemChannelBalanceLimitUpdate(req *types.SystemChannelBalanceLimitUpdateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("bs_system_params").Where("name = 'channelBalanceLimit'").Update("value", req.BalanceLimit).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}
