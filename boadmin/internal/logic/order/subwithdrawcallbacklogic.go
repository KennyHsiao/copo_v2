package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"go.opentelemetry.io/otel/trace"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubWithdrawCallBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubWithdrawCallBackLogic(ctx context.Context, svcCtx *svc.ServiceContext) SubWithdrawCallBackLogic {
	return SubWithdrawCallBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubWithdrawCallBackLogic) SubWithdrawCallBack(req *types.ProxyPayOrderCallBackRequest) (resp *types.ProxyPayOrderCallBackResponse, err error) {
	logx.WithContext(l.ctx).Infof("渠道回調請求參數: %+v", req)
	//检查单号是否存在
	orderX := &types.OrderX{}
	if req.ProxyPayOrderNo == "" && req.ChannelOrderNo == "" {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST)
	} else if orderX, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.ProxyPayOrderNo, ""); err != nil && orderX == nil {
		return nil, errorz.New(response.ORDER_NUMBER_NOT_EXIST, "Copo OrderNo: "+req.ProxyPayOrderNo)
	}

	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
		MerchantCode:    orderX.MerchantCode,
		MerchantOrderNo: orderX.MerchantOrderNo,
		OrderNo:         orderX.OrderNo,
		LogType:         constants.CALLBACK_FROM_CHANNEL,
		LogSource:       constants.API_DF,
		Content:         req,
		TraceId:         trace.SpanContextFromContext(l.ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", errLog)
	}

	// todo: 更改訂單狀態

	return
}
