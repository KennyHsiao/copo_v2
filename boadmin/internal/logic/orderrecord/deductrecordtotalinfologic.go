package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductRecordTotalInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeductRecordTotalInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) DeductRecordTotalInfoLogic {
	return DeductRecordTotalInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeductRecordTotalInfoLogic) DeductRecordTotalInfo(req types.DeductRecordQueryAllRequestX) (resp *types.DeductRecordTotalInfoResponse, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}

	resp, err = orderrecordService.DeductRecordTotalInfo(l.svcCtx.MyDB, req, l.ctx)

	return
}
