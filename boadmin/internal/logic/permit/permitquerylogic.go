package permit

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitQueryLogic {
	return PermitQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitQueryLogic) PermitQuery(req types.PermitQueryRequest) (resp *types.PermitQueryResponse, err error) {
	err = l.svcCtx.MyDB.Table("au_permits").Take(&resp, req.ID).Error
	return
}
