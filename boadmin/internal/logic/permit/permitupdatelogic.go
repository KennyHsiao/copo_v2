package permit

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitUpdateLogic {
	return PermitUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitUpdateLogic) PermitUpdate(req types.PermitUpdateRequest) error {
	permit := &types.PermitUpdate{
		PermitUpdateRequest: req,
	}
	return l.svcCtx.MyDB.Table("au_permits").Updates(permit).Error
}
