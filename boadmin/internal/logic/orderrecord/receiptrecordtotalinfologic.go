package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReceiptRecordTotalInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReceiptRecordTotalInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReceiptRecordTotalInfoLogic {
	return ReceiptRecordTotalInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReceiptRecordTotalInfoLogic) ReceiptRecordTotalInfo(req types.ReceiptRecordQueryAllRequestX) (resp *types.ReceiptRecordTotalInfoResponse, err error) {
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}
	var orderAmount float64
	resp, err = orderrecordService.ReceiptRecordTotalInfoBySuccess(l.svcCtx.MyDB, req, l.ctx)
	orderAmount, err = orderrecordService.ReceiptRecordTotalOrderAmount(l.svcCtx.MyDB, req, l.ctx)
	resp.TotalOrderAmount = orderAmount
	return
}
