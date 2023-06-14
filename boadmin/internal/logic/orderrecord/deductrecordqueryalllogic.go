package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeductRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeductRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) DeductRecordQueryAllLogic {
	return DeductRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeductRecordQueryAllLogic) DeductRecordQueryAll(req types.DeductRecordQueryAllRequestX) (resp *types.DeductRecordQueryAllResponseX, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}
	resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, false, l.ctx)
	return
}
