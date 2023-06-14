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

type SystemIpUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemIpUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemIpUpdateLogic {
	return SystemIpUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemIpUpdateLogic) SystemIpUpdate(req *types.SystemIpRequest) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("bs_system_params").Where("name = 'managerIPWhiteList'").Update("value", req.IpString).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}
