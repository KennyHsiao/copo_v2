package order

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"fmt"
	"github.com/neccoys/go-zero-extension/redislock"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubWithdrawReviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubWithdrawReviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) SubWithdrawReviewLogic {
	return SubWithdrawReviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubWithdrawReviewLogic) SubWithdrawReview(req *types.SubWithdrawOrderUpdateRequest) (resp string, err error) {
	//JWT取得登入腳色資訊 用於商戶號与使用者账号
	userAccount := l.ctx.Value("account").(string)
	order := types.OrderX{}
	var channelWithdraws []types.TxOrderChannels
	var merchantBalanceRecord types.MerchantBalanceRecord

	// 若为不通过，必须要写理由
	status := req.Status
	if len(status) > 0 && status == constants.FAIL {
		if len(req.Memo) < 0 {
			return "", errorz.New(response.REVIEW_REASON_ERROR)
		}
	} else if len(status) > 0 && status == constants.SUCCESS {
	}

	// TODO 是否需要在检查一次 order orderAmount == 子单成功状态 金额加总??

	// 判断下发单号不可为空
	if &req.OrderNo == nil { // 订单号不得为空
		return "", errorz.New(response.INVALID_WITHDRAW_ORDER_NO)
	}

	if err := l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", &req.OrderNo).Find(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errorz.New(response.DATA_NOT_FOUND)
		}
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if order.Status == constants.SUCCESS || order.Status == constants.FAIL { // 已結單的下發訂單不可重複審核
		return "", errorz.New(response.COMPLETED_ORDER_REVIVEW_REPEAT)
	}

	//更新子单transferhandlingfee1
	if err = l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		if status == constants.SUCCESS {

			//更新tx_order tx_order_channel transferhandlingfee / 总子单数目
			if err := l.svcCtx.MyDB.Table("tx_order_channels").
				Where("status = '20' AND order_no = ?", req.OrderNo). //訂單狀態(0:待處理 1:處理中 2:交易中  20:成功 30:失敗 31:凍結)
				Find(&channelWithdraws).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE)
			} else if len(channelWithdraws) == 0 {
				return errorz.New(response.INVALID_REVIEW_SUCCESS)
			}
			//計算商戶手續費平均在每個渠道
			perChannelWithdrawHandlingFee := utils.FloatDiv(order.HandlingFee, l.intToFloat64(len(channelWithdraws)))

			if err := l.svcCtx.MyDB.Table("tx_order_channels").
				Where("status = '20' AND order_no = ?", req.OrderNo).
				Updates(map[string]interface{}{"transfer_handling_fee": perChannelWithdrawHandlingFee}).Error; err != nil {
				return errors.New(response.DATABASE_FAILURE)
			}

			//更新tx_order status 20 成功 sub_withdraw_all_status: 1 ,review_by : UserAccount , memo
			order.TransAt = types.JsonTime{}.New()
			order.Status = constants.SUCCESS
			order.ReviewedBy = userAccount
			order.Memo = req.Memo

			if err := l.svcCtx.MyDB.Table("tx_orders").
				Updates(order).Error; err != nil {
				return errors.New(response.DATABASE_FAILURE)
			}

			var totaTransferHandlingFeelFee float64
			var totalHandlingFee float64
			var totalFee float64
			for _, channelWithdraw := range channelWithdraws {
				channelPayType := &types.ChannelPayType{}
				if err1 := l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("code = ?", channelWithdraw.ChannelCode+"DF").
					Take(channelPayType).Error; err1 != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errorz.New(response.DATA_NOT_FOUND)
					}
					return errorz.New(response.DATABASE_FAILURE, err1.Error())
				}
				totalHandlingFee = utils.FloatAdd(totalHandlingFee, channelPayType.HandlingFee)                       //總手續費
				totaTransferHandlingFeelFee = utils.FloatAdd(totaTransferHandlingFeelFee, channelPayType.HandlingFee) //總手續費 + 總費率計算
				if channelPayType.IsRate == "1" {
					totalFee = utils.FloatAdd(totalFee, utils.FloatDiv(utils.FloatMul(channelWithdraw.OrderAmount, channelPayType.Fee), 100)) //總費率計算 (金额X费率)
					totaTransferHandlingFeelFee = utils.FloatAdd(totaTransferHandlingFeelFee, utils.FloatDiv(utils.FloatMul(channelWithdraw.OrderAmount, channelPayType.Fee), 100))
				}
			}

			if err = l.CalculateSystemProfit(db, &order, totaTransferHandlingFeelFee, totalHandlingFee, totalFee); err != nil {
				logx.WithContext(l.ctx).Errorf("审核通过，计算下发利润失败, 订单号: %s, err : %s", order.OrderNo, err.Error())
				return err
			}

			// 新單新增訂單歷程 (不抱錯)
			if err4 := l.svcCtx.MyDB.Table("tx_order_actions").Create(&types.OrderActionX{
				OrderAction: types.OrderAction{
					OrderNo:     order.OrderNo,
					Action:      "REVIEW_SUCCESS",
					UserAccount: userAccount,
					Comment:     order.Memo,
				},
			}).Error; err4 != nil {
				logx.WithContext(l.ctx).Error("紀錄訂單歷程出錯:%s", err4.Error())
			}

		} else if status == constants.FAIL {
			// 異動錢包
			if merchantBalanceRecord, err = l.UpdateBalance(db, types.UpdateBalance{
				MerchantCode:    order.MerchantCode,
				CurrencyCode:    order.CurrencyCode,
				OrderNo:         order.OrderNo,
				OrderType:       order.Type,
				TransactionType: "4",
				BalanceType:     order.BalanceType,
				TransferAmount:  order.TransferAmount,
				Comment:         order.Memo,
				CreatedBy:       userAccount,
			}); err != nil {
				return err
			}

			order.BeforeBalance = merchantBalanceRecord.BeforeBalance
			order.Balance = merchantBalanceRecord.AfterBalance
			order.TransAt = types.JsonTime{}.New()

			order.Status = constants.FAIL
			order.ReviewedBy = userAccount
			order.Memo = req.Memo

			if err := l.svcCtx.MyDB.Table("tx_orders").
				Updates(order).Error; err != nil {
				return errors.New(response.DATABASE_FAILURE)
			}

			// 新單新增訂單歷程 (不抱錯)
			if err4 := l.svcCtx.MyDB.Table("tx_order_actions").Create(&types.OrderActionX{
				OrderAction: types.OrderAction{
					OrderNo:     order.OrderNo,
					Action:      "REVIEW_FAIL",
					UserAccount: userAccount,
					Comment:     req.Memo,
				},
			}).Error; err4 != nil {
				logx.WithContext(l.ctx).Error("紀錄訂單歷程出錯:%s", err4.Error())
			}
		}

		return
	}); err != nil {
		return "下发审核更新失敗，orderNo = " + req.OrderNo + "，err : " + err.Error(), err
	}

	return "操作成功", nil
}

// CalculateSystemProfit 計算系統利潤
func (l *SubWithdrawReviewLogic) CalculateSystemProfit(db *gorm.DB, order *types.OrderX, TotalChannelHandlingFee float64, totalHandlingFee float64, totalFee float64) (err error) {

	systemFeeProfit := types.OrderFeeProfit{
		OrderNo:             order.OrderNo,
		MerchantCode:        "00000000",
		BalanceType:         order.BalanceType,
		Fee:                 totalFee,
		HandlingFee:         totalHandlingFee,
		TransferHandlingFee: TotalChannelHandlingFee,
		// 商戶手續費 - 渠道總手續費 = 利潤 (有可能是負的)
		ProfitAmount: utils.FloatSub(order.TransferHandlingFee, TotalChannelHandlingFee),
	}

	// 保存系統利潤
	if err = db.Table("tx_orders_fee_profit").Create(&types.OrderFeeProfitX{
		OrderFeeProfit: systemFeeProfit,
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l SubWithdrawReviewLogic) intToFloat64(i int) float64 {
	intStr := strconv.Itoa(i)
	res, _ := strconv.ParseFloat(intStr, 64)
	return res
}

func (l SubWithdrawReviewLogic) UpdateBalance(db *gorm.DB, updateBalance types.UpdateBalance) (merchantBalanceRecord types.MerchantBalanceRecord, err error) {
	redisKey := fmt.Sprintf("%s-%s-%s", updateBalance.MerchantCode, updateBalance.CurrencyCode, updateBalance.BalanceType)
	redisLock := redislock.New(l.svcCtx.RedisClient, redisKey, "merchant-balance:")
	redisLock.SetExpire(5)
	if isOk, _ := redisLock.TryLockTimeout(5); isOk {
		defer redisLock.Release()
		if merchantBalanceRecord, err = l.doUpdateBalance(db, updateBalance); err != nil {
			return
		}
	} else {
		return merchantBalanceRecord, errorz.New(response.BALANCE_REDISLOCK_ERROR)
	}
	return
}

func (l SubWithdrawReviewLogic) doUpdateBalance(db *gorm.DB, updateBalance types.UpdateBalance) (merchantBalanceRecord types.MerchantBalanceRecord, err error) {
	var beforeBalance float64
	var afterBalance float64

	// 1. 取得 商戶餘額表
	var merchantBalance types.MerchantBalance
	if err = db.Table("mc_merchant_balances").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("merchant_code = ? AND currency_code = ? AND balance_type = ?", updateBalance.MerchantCode, updateBalance.CurrencyCode, updateBalance.BalanceType).
		Take(&merchantBalance).Error; err != nil {
		return merchantBalanceRecord, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 2. 計算
	var selectBalance string
	if utils.FloatAdd(merchantBalance.Balance, updateBalance.TransferAmount) < 0 {
		logx.WithContext(l.ctx).Errorf("商户:%s，余额类型:%s，余额:%s，交易金额:%s", merchantBalance.MerchantCode, merchantBalance.BalanceType, fmt.Sprintf("%f", merchantBalance.Balance), fmt.Sprintf("%f", updateBalance.TransferAmount))
		return merchantBalanceRecord, errorz.New(response.MERCHANT_INSUFFICIENT_DF_BALANCE)
	}
	selectBalance = "balance"
	beforeBalance = merchantBalance.Balance
	afterBalance = utils.FloatAdd(beforeBalance, updateBalance.TransferAmount)
	merchantBalance.Balance = afterBalance

	// 3. 變更 商戶餘額
	if err = db.Table("mc_merchant_balances").Select(selectBalance).Updates(types.MerchantBalanceX{
		MerchantBalance: merchantBalance,
	}).Error; err != nil {
		logx.WithContext(l.ctx).Error(err.Error())
		return merchantBalanceRecord, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 4. 新增 餘額紀錄
	merchantBalanceRecord = types.MerchantBalanceRecord{
		MerchantBalanceId: merchantBalance.ID,
		MerchantCode:      merchantBalance.MerchantCode,
		CurrencyCode:      merchantBalance.CurrencyCode,
		OrderNo:           updateBalance.OrderNo,
		OrderType:         updateBalance.OrderType,
		ChannelCode:       updateBalance.ChannelCode,
		PayTypeCode:       updateBalance.PayTypeCode,
		TransactionType:   updateBalance.TransactionType,
		BalanceType:       updateBalance.BalanceType,
		BeforeBalance:     beforeBalance,
		TransferAmount:    updateBalance.TransferAmount,
		AfterBalance:      afterBalance,
		Comment:           updateBalance.Comment,
		CreatedBy:         updateBalance.CreatedBy,
	}

	if err = db.Table("mc_merchant_balance_records").Create(&types.MerchantBalanceRecordX{
		MerchantBalanceRecord: merchantBalanceRecord,
	}).Error; err != nil {
		return merchantBalanceRecord, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
