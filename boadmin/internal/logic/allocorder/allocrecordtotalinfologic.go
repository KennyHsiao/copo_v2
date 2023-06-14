package allocorder

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AllocRecordTotalInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAllocRecordTotalInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) AllocRecordTotalInfoLogic {
	return AllocRecordTotalInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AllocRecordTotalInfoLogic) AllocRecordTotalInfo(req types.AllocRecordQueryAllRequestX) (resp *types.AllocRecordTotalInfoResponse, err error) {
	resp, err = orderrecordService.AllocRecordTotalInfo(l.svcCtx.MyDB, req)
	return
}
