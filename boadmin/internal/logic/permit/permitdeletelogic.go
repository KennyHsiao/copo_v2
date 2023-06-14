package permit

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitDeleteLogic {
	return PermitDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitDeleteLogic) PermitDelete(req types.PermitDeleteRequest) error {
	return l.svcCtx.MyDB.Table("au_permits").Delete(&req).Error
}
