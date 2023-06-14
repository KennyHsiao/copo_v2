package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReceiptRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReceiptRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReceiptRecordQueryAllLogic {
	return ReceiptRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReceiptRecordQueryAllLogic) ReceiptRecordQueryAll(req types.ReceiptRecordQueryAllRequestX) (resp *types.ReceiptRecordQueryAllResponseX, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}
	resp, err = orderrecordService.ReceiptRecordQueryAll(l.svcCtx.MyDB, req, false, l.ctx)
	for i, record := range resp.List {
		if len(record.InternalChargeOrderPath) > 0 {
			resp.List[i].InternalChargeOrderPath = l.svcCtx.Config.ResourceHost + record.InternalChargeOrderPath
		}
	}

	return
}
