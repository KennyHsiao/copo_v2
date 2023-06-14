package order

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type WithdrawReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawReviewLogic {
	return WithdrawReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawReviewLogic) WithdrawReview(req types.WithdrawOrderUpdateRequest) error {
	//JWT取得登入腳色資訊 用於商戶號与使用者账号
	userAccount := l.ctx.Value("account").(string)
	order := types.OrderX{}

	// 若为不通过，必须要写理由
	status := req.Status
	if len(status) > 0 && status == constants.FAIL {
		if len(req.Memo) < 0 {
			return errorz.New(response.REVIEW_REASON_ERROR)
		}
	}

	// 判断下发单号不可为空
	if &req.OrderNo == nil { // 订单号不得为空
		return errorz.New(response.INVALID_WITHDRAW_ORDER_NO)
	}

	if err := l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", &req.OrderNo).Find(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND)
		}
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if order.Status == constants.SUCCESS || order.Status == constants.FAIL { // 已結單的下發訂單不可重複審核
		return errorz.New(response.COMPLETED_ORDER_REVIVEW_REPEAT)
	}

	if status == constants.SUCCESS { // 审核通过
		// 判断是否有选渠道
		channelWithdraws := req.List
		if channelWithdraws == nil || len(channelWithdraws) == 0 {
			return errorz.New(response.MERCHANT_WITHDRAW_RECORD_ERROR)
		}

		var errRpc error
		var res *transaction.WithdrawReviewSuccessResponse
		var rpcChannelWithdraws []*transaction.ChannelWithdraw
		rpcChannelWithdraws = generateRpcData(channelWithdraws)
		res, errRpc = l.svcCtx.TransactionRpc.WithdrawReviewSuccessTransaction(l.ctx, &transaction.WithdrawReviewSuccessRequest{
			OrderNo:         order.OrderNo,
			UserAccount:     userAccount,
			Memo:            req.Memo,
			ChannelWithdraw: rpcChannelWithdraws,
		})

		if errRpc != nil {
			logx.WithContext(l.ctx).Error("WithdrawReviewPass Tranaction rpcResp error:%s", errRpc.Error())
			return errorz.New(response.FAIL, errRpc.Error())
		} else if res.Code != response.API_SUCCESS {
			logx.WithContext(l.ctx).Errorf("WithdrawReviewPass Tranaction error Code:%s, Message:%s", res.Code, res.Message)
			return errorz.New(res.Code, res.Message)
		} else if res.Code == response.API_SUCCESS {
			logx.WithContext(l.ctx).Infof("下发审核通过rpc完成，单号: %#v", res.OrderNo)
		}
	} else if req.Status == constants.FAIL { // 审核不通过
		var errRpc error
		var res *transaction.WithdrawReviewFailResponse
		//var merchantChannelRate types.MerchantChannelRate
		//
		//if err := l.svcCtx.MyDB.Table("mc_merchant_channel_rate").
		//	Where("merchant_code = ?", order.MerchantCode).
		//	Where("currency_code = ?", req.CurrencyCode).
		//	Where("channel_pay_types_code = ?", req.ChannelPayTypesCode).Find(&merchantChannelRate).Error; err != nil {
		//		return errorz.New(response.DATABASE_FAILURE, err.Error())
		//}

		res, errRpc = l.svcCtx.TransactionRpc.WithdrawReviewFailTransaction(l.ctx, &transaction.WithdrawReviewFailRequest{
			OrderNo:     order.OrderNo,
			UserAccount: userAccount,
			Memo:        req.Memo,
			PtBalanceId: req.PtBalanceId,
		})

		if errRpc != nil {
			logx.WithContext(l.ctx).Error("WithdrawReviewNotPass Tranaction rpcResp error:%s", errRpc.Error())
			return errorz.New(response.FAIL, errRpc.Error())
		} else if res.Code != response.API_SUCCESS {
			logx.WithContext(l.ctx).Errorf("WithdrawReviewNotPass Tranaction error Code:%s, Message:%s", res.Code, res.Message)
			return errorz.New(res.Code, res.Message)
		} else if res.Code == response.API_SUCCESS {
			logx.WithContext(l.ctx).Infof("下发审核不通过rpc完成，单号: %#v", res.OrderNo)
		}
	}

	// 回调商户
	if order.Source == constants.API && len(order.NotifyUrl) > 0 {
		// 异步回调
		go func() {
			l.callNoticeURL(order)
		}()
	}

	return nil
}

func generateRpcData(withdraws []types.ChannelWithdraw) []*transaction.ChannelWithdraw {

	var resp []*transaction.ChannelWithdraw

	for i, _ := range withdraws {
		resp = append(resp, &transaction.ChannelWithdraw{
			ChannelCode:    withdraws[i].ChannelCode,
			WithdrawAmount: withdraws[i].WithdrawAmount,
		})
	}

	return resp
}

// CalculateSystemProfit 計算系統利潤
func (l *WithdrawReviewLogic) CalculateSystemProfit(db *gorm.DB, order *types.Order, TransferHandlingFee float64) (err error) {

	systemFeeProfit := types.OrderFeeProfit{
		OrderNo:             order.OrderNo,
		MerchantCode:        "00000000",
		BalanceType:         order.BalanceType,
		Fee:                 0,
		HandlingFee:         TransferHandlingFee,
		TransferHandlingFee: TransferHandlingFee,
		// 商戶手續費 - 渠道總手續費 = 利潤 (有可能是負的)
		ProfitAmount: utils.FloatSub(order.TransferHandlingFee, TransferHandlingFee),
	}

	// 保存系統利潤
	if err = db.Table("tx_orders_fee_profit").Create(&types.OrderFeeProfitX{
		OrderFeeProfit: systemFeeProfit,
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l *WithdrawReviewLogic) intToFloat64(i int) float64 {
	intStr := strconv.Itoa(i)
	res, _ := strconv.ParseFloat(intStr, 64)
	return res
}

func (l *WithdrawReviewLogic) callNoticeURL(order types.OrderX) error {
	var minDelaySeconds int64 = 10
	for i := 0; i < 5; i++ {
		startTime := time.Now().Unix()
		logx.WithContext(l.ctx).Infof("PayCallback To Merchant: 第%d次回調 訂單:%s, NotifyUrl:%s", i+1, order.OrderNo, order.NotifyUrl)
		if len(order.ChangeType) > 0 && order.ChangeType == "1" {
			if err := merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, &order); err != nil {
				logx.WithContext(l.ctx).Error("下发回調商戶錯誤(代付参数):", err)
			} else {
				break
			}
		} else {
			err := ordersService.WithdrawApiCallBack(l.svcCtx.MyDB, order, l.ctx)
			if err != nil {
				logx.WithContext(l.ctx).Error("下发回調商戶錯誤:", err)
			} else {
				break
			}
		}
		endTime := time.Now().Unix()
		secondsDiff := endTime - startTime
		if secondsDiff < minDelaySeconds {
			sleepTime := time.Duration(minDelaySeconds-secondsDiff) * time.Second
			time.Sleep(sleepTime)
		}
	}
	return nil
}
