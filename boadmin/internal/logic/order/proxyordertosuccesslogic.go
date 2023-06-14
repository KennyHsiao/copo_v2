package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantbalanceservice"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyOrderToSuccessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyOrderToSuccessLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyOrderToSuccessLogic {
	return ProxyOrderToSuccessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyOrderToSuccessLogic) ProxyOrderToSuccess(req *types.ProxyOrderToSuccessRequest) (resp *types.ProxyOrderToSuccessResponse, err error) {

	txOrder := &types.OrderX{}
	if txOrder, err = model.QueryOrderByOrderNo(l.svcCtx.MyDB, req.OrderNo, ""); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if txOrder.Status != constants.TRANSACTION && txOrder.Status != constants.WAIT_PROCESS {
		return nil, errorz.New(response.ORDER_STATUS_IS_NOT_TRANSACTION)
	}

	//JWT取得登入账号
	account := l.ctx.Value("account").(string)
	txOrder.UpdatedBy = account
	txOrder.UpdatedAt = time.Now().UTC()
	txOrder.ChannelCallBackAt = time.Now().UTC()
	txOrder.TransAt = types.JsonTime{}.New()
	txOrder.RepaymentStatus = constants.REPAYMENT_NOT //还款状态：([0]：不需还款、1：待还款、2：还款成功、3：还款失败)，预设不需还款

	var orderAction string
	var common string
	if req.Status == "20" {
		if txOrder.Status == constants.WAIT_PROCESS {
			return nil, errorz.New(response.PROCESSING_ORDER_NOT_TO_SUCCESS)
		}
		txOrder.Status = constants.SUCCESS
		txOrder.CallBackStatus = "1" //此处固定写成功"1"，预设渠道那端是成功单
		txOrder.Memo = "人工调整成功单"
		orderAction = "PERSON_SUCCESS"
		common = "人工调整成功单"
	} else if req.Status == "30" {
		txOrder.Status = constants.FAIL
		txOrder.CallBackStatus = "2" //此处固定写成功"2"，预设渠道那端是失敗单
		txOrder.Memo = "人工调整失敗单"
		orderAction = "PERSON_FAILURE"
		common = "人工调整失敗单"

		txOrder.RepaymentStatus = constants.REPAYMENT_WAIT //还款状态：(0：不需还款、[1]：待还款、2：还款成功、3：还款失败)，预设不需还款
		txOrder.ErrorNote = "渠道回调: 交易失败"

	}

	// 更新订单
	if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(txOrder).Error; errUpdate != nil {
		logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
	}

	// 更新訂單訂單歷程 (不抱錯)
	if err4 := l.svcCtx.MyDB.Table("tx_order_actions").Create(&types.OrderActionX{
		OrderAction: types.OrderAction{
			OrderNo:     txOrder.OrderNo,
			Action:      orderAction,
			UserAccount: txOrder.MerchantCode,
			Comment:     common,
		},
	}).Error; err4 != nil {
		logx.Error("紀錄訂單歷程出錯:%s", err4.Error())
	}

	if req.Status == "30" {
		logx.WithContext(l.ctx).Info("代付订单回调状态为[失败]，开始还款=======================================>", txOrder.Order.OrderNo)
		//呼叫RPC
		balanceType, errBalance := merchantbalanceservice.GetBalanceTypeByOrder(l.svcCtx.MyDB, txOrder.OrderNo)
		if errBalance != nil {
			return nil, errBalance
		}

		var errRpc error
		//當訂單還款狀態為待还款
		if txOrder.RepaymentStatus == constants.REPAYMENT_WAIT {
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			if balanceType == "DFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: txOrder.MerchantCode,
					OrderNo:      txOrder.OrderNo,
				})
			} else if balanceType == "XFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: txOrder.MerchantCode,
					OrderNo:      txOrder.OrderNo,
				})
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Errorf("代付提单回调 %s 还款失败。 Err: %s", txOrder.OrderNo, errRpc.Error())
				txOrder.RepaymentStatus = constants.REPAYMENT_FAIL

				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(txOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}

				return nil, errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.WithContext(l.ctx).Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				txOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(txOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}
			}
		}
	}

	return
}
