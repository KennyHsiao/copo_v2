package allocorder

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AllocRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAllocRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) AllocRecordQueryAllLogic {
	return AllocRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AllocRecordQueryAllLogic) AllocRecordQueryAll(req types.AllocRecordQueryAllRequestX) (resp *types.AllocRecordQueryAllResponseX, err error) {
	resp, err = orderrecordService.AllocRecordQueryAll(l.svcCtx.MyDB, req)
	return
}
