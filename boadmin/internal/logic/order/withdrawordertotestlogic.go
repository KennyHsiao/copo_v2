package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/copo888/transaction_service/rpc/transactionclient"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderToTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderToTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderToTestLogic {
	return WithdrawOrderToTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawOrderToTestLogic) WithdrawOrderToTest(req *types.WithdrawOrderToTestRequest) (resp *types.WithdrawOrderToTestResponse, err error) {
	txOrder := &types.OrderX{}
	if txOrder, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.OrderNo, ""); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if txOrder.Status != constants.SUCCESS {
		return nil, errorz.New(response.ORDER_STATUS_IS_NOT_FOR_TEST)
	}

	resRpc := &transaction.WithdrawOrderTestResponse{}
	var errRpc error

	if txOrder.IsTest == "1" { //從測試單改正式單
		resRpc, errRpc = l.svcCtx.TransactionRpc.WithdrawTestToNormal_XFB(l.ctx, &transactionclient.WithdrawOrderTestRequest{
			WithdrawOrderNo: txOrder.OrderNo,
			PtBalanceId:     req.PtBalanceId,
			Remark:          req.Remark,
		})

		if errRpc != nil {
			logx.Errorf("下發转正式 %s 还款失败。 Err: %s", txOrder.OrderNo, errRpc.Error())
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else if resRpc.Code != "000" {
			logx.Errorf("下發转正式 %s 失败。 ErrCode: %s, ErrMsg: %s", txOrder.OrderNo, resRpc.Code, resRpc.Message)
			resp = &types.WithdrawOrderToTestResponse{
				RespCode: resRpc.Code,
				RespMsg:  resRpc.Message,
			}
			err = errorz.New(response.WALLET_UPDATE_ERROR, resRpc.Message)
			return
		} else {
			logx.Infof("下發還款rpc完成，%s 錢包還款完成: %#v", txOrder.BalanceType, resRpc)
		}
	} else if txOrder.IsTest == "0" { //從正式單改測試單
		//如果"已結算"/"確認報表無誤按鈕" : 不扣款月結傭金錢包

		//1. 更新測試單flag
		//2. 将商户钱包加回 (orderNO)
		//3. 更新錢包紀錄
		resRpc, errRpc = l.svcCtx.TransactionRpc.WithdrawOrderToTest_XFB(l.ctx, &transactionclient.WithdrawOrderTestRequest{
			WithdrawOrderNo: txOrder.OrderNo,
			PtBalanceId:     req.PtBalanceId,
			Remark:          req.Remark,
		})

		if errRpc != nil {
			logx.Errorf("下發转测试 %s 失败。 Err: %s", txOrder.OrderNo, errRpc.Error())
			return nil, errorz.New(response.FAIL, errRpc.Error())
		} else if resRpc.Code != "000" {
			logx.Errorf("下發转测试 %s 失败。 ErrCode: %s, ErrMsg: %s", txOrder.OrderNo, resRpc.Code, resRpc.Message)
			resp = &types.WithdrawOrderToTestResponse{
				RespCode: resRpc.Code,
				RespMsg:  resRpc.Message,
			}
			err = errorz.New(response.WALLET_UPDATE_ERROR, resRpc.Message)
			return
		} else {
			logx.Infof("下發還款rpc完成，%s 錢包還款完成: %#v", txOrder.BalanceType, resRpc)
		}
	}

	return
}
