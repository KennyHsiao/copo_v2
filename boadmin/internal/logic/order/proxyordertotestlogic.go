package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyOrderToTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyOrderToTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyOrderToTestLogic {
	return ProxyOrderToTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyOrderToTestLogic) ProxyOrderToTest(req *types.ProxyOrderToTestRequest) (resp *types.ProxyOrderToTestResponse, err error) {

	txOrder := &types.OrderX{}
	if txOrder, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.OrderNo, ""); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if txOrder.Status != constants.SUCCESS {
		return nil, errorz.New(response.ORDER_STATUS_IS_NOT_FOR_TEST)
	}

	var resRpc *transaction.ProxyOrderTestResponse
	var errRpc error

	if txOrder.IsTest == "1" { //從測試單改正式單
		if txOrder.BalanceType == "DFB" {
			resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyTestToNormal_DFB(l.ctx, &transactionclient.ProxyOrderTestRequest{
				ProxyOrderNo: txOrder.OrderNo,
			})
		} else if txOrder.BalanceType == "XFB" {
			resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyTestToNormal_XFB(l.ctx, &transactionclient.ProxyOrderTestRequest{
				ProxyOrderNo: txOrder.OrderNo,
			})
		}

		if errRpc != nil {
			logx.Errorf("代付提单 %s 还款失败。 Err: %s", txOrder.OrderNo, errRpc.Error())
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else {
			logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", txOrder.BalanceType, resRpc)
		}

	} else if txOrder.IsTest == "0" { //從正式單改測試單
		//如果"已結算"/"確認報表無誤按鈕" : 不扣款月結傭金錢包

		//1. 更新測試單flag
		//2. 将商户钱包加回 (orderNO)
		//3. 更新錢包紀錄
		if txOrder.BalanceType == "DFB" {
			resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderToTest_DFB(l.ctx, &transactionclient.ProxyOrderTestRequest{
				ProxyOrderNo: txOrder.OrderNo,
			})
		} else if txOrder.BalanceType == "XFB" {
			resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderToTest_XFB(l.ctx, &transactionclient.ProxyOrderTestRequest{
				ProxyOrderNo: txOrder.OrderNo,
			})
		}

		if errRpc != nil {
			logx.Errorf("代付提单 %s 还款失败。 Err: %s", txOrder.OrderNo, errRpc.Error())
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else {
			logx.Infof("代付還款rpc完成，%s 錢包還款完成: %#v", txOrder.BalanceType, resRpc)
		}

	}

	return
}
