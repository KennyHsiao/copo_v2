package permit

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitCreateLogic {
	return PermitCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitCreateLogic) PermitCreate(req types.PermitCreateRequest) error {
	permit := &types.PermitCreate{
		PermitCreateRequest: req,
	}
	return l.svcCtx.MyDB.Table("au_permits").Create(permit).Error
}
