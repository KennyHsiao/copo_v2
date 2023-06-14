package order

import (
	"com.copo/bo_service/boadmin/internal/model"
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
)

type OrderUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderUpdateLogic {
	return OrderUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderUpdateLogic) OrderUpdate(req types.OrderUpdateRequest) error {
	// JWT取得登入腳色資訊 用於商戶號与账号
	userAccount := l.ctx.Value("account").(string)
	// 判断是否有这张单
	order := types.OrderX{}

	if err := l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", req.OrderNo).
		Where("currency_code = ?", req.CurrencyCode).
		Take(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND, err.Error())
		}
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 确认是否为處理中
	if order.Status != "1" { // 已結單的下發訂單不可重複審核
		return errorz.New(response.COMPLETED_ORDER_REVIVEW_REPEAT)
	}

	// 取得商户检查费率状态
	var merchant types.Merchant
	if err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", order.MerchantCode).Take(&merchant).Error; err != nil {
		return errorz.New(response.DATA_NOT_FOUND, err.Error())
	}

	// 审核通过
	if req.Status == "20" {
		channel := types.ChannelData{}
		if err := l.svcCtx.MyDB.Table("ch_channels").Where("code = ?", req.ChannelCode).
			Where("currency_code = ?", req.CurrencyCode).
			Take(&channel).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorz.New(response.DATA_NOT_FOUND, err.Error())
			}
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		ptBalanceId, err := l.updateOrderByChannelIsProxy(&order, &req, channel.IsProxy, merchant.RateCheck, merchant.Code)
		if err != nil {
			return err
		}

		var errRpc error
		var res *transaction.InternalReviewSuccessResponse
		var rates []*transaction.Rates
		for _, rate := range req.NcRates {
			rate := &transaction.Rates{
				AgentLayerCode: rate.AgentlayerCode,
				Rate:           rate.Rate,
			}
			rates = append(rates, rate)
		}
		res, errRpc = l.svcCtx.TransactionRpc.InternalReviewSuccessTransaction(l.ctx, &transaction.InternalReviewSuccessRequest{
			OrderNo:     order.OrderNo,
			UserAccount: userAccount,
			ChnRate:     req.ChnRate,
			List:        rates,
			IsProxy:     channel.IsProxy,
			PtBalanceId: ptBalanceId,
		})

		if errRpc != nil {
			logx.Error("InternalChargeReview Tranaction rpcResp error:%s", errRpc.Error())
			return errorz.New(response.FAIL, errRpc.Error())
		} else if res.Code != response.API_SUCCESS {
			logx.Errorf("InternalChargeReview Tranaction error Code:%s, Message:%s", res.Code, res.Message)
			return errorz.New(res.Code, res.Message)
		} else if res.Code == response.API_SUCCESS {
			logx.Infof("内充审核通过rpc完成，%s 錢包充值完成: %#v", "DFB", res.OrderNo)
		}
	} else if req.Status == "30" { // 审核不通过
		// 若为审核不通过，备注必须有value
		if len(req.Memo) < 0 {
			return errorz.New(response.REVIEW_REASON_ERROR)
		}
		// 交易時間
		order.TransAt = types.JsonTime{}.New()
		order.Status = constants.FAIL
		order.Memo = req.Memo
		return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
			// 更新tx_order
			if err := l.svcCtx.MyDB.Table("tx_orders").Where("currency_code = ?", req.CurrencyCode).Updates(order).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			//记录订单历程
			orderAction := types.OrderAction{
				OrderNo:     order.OrderNo,
				Action:      "REVIEW_FAIL",
				UserAccount: userAccount,
				Comment:     req.Memo,
			}
			if err := model.NewOrderAction(db).CreateOrderAction(&types.OrderActionX{
				OrderAction: orderAction,
			}); err != nil {
				return err
			}
			return nil
		})

	}

	return nil
}

func (l OrderUpdateLogic) compareRates(rates []types.InternalChargeRate, merRate map[string]string) (isRateOk, isProxyOk bool) {
	if len(merRate) != len(rates) {
		return false, false
	}

	for i := 0; i < len(rates); i++ {
		k := i + 1
		if k > (len(rates) - 1) {
			break
		}
		_, isExist := merRate[rates[i].AgentlayerCode]
		if !isExist {
			return false, false
		}
		_, isExist2 := merRate[rates[k].AgentlayerCode]
		if !isExist2 {
			return false, false
		}
		com := utils.FloatSub(rates[i].Rate, rates[k].Rate)
		if com > 0 {
			return false, true
		}
	}

	return true, true
}

func (l OrderUpdateLogic) getAllParentMerchant(merchantCode string, merMap map[string]string) (resp map[string]string, err error) {

	merchant, errM := l.getMerchantInfo(merchantCode)
	if errM != nil {
		return nil, errM
	}
	merMap[merchant.AgentLayerCode] = merchant.Code
	if len(merchant.AgentParentCode) > 0 {
		l.getAllParentMerchant(merchant.AgentParentCode, merMap)
	}
	resp = merMap
	return resp, err
}

func (l OrderUpdateLogic) getMerchantInfo(merchantCode string) (resp *types.Merchant, err error) {
	var merchant types.Merchant
	if err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
		return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
	}
	return &merchant, nil
}

func (l OrderUpdateLogic) updateOrderByChannelIsProxy(order *types.OrderX, req *types.OrderUpdateRequest, isProxy, rateCheck, merchantCode string) (ptBalanceId int64, err error) {
	if isProxy != constants.ISPROXY {

		merchantOrderRateListView, err3 := ordersService.GetMerchantChannelRate(l.svcCtx.MyDB, order.MerchantCode, order.CurrencyCode, constants.ORDER_TYPE_NC)
		if err3 != nil {
			return 0, err3
		}
		if rateCheck != "0" {
			if merchantOrderRateListView[0].CptHandlingFee > merchantOrderRateListView[0].MerHandlingFee || merchantOrderRateListView[0].CptFee > merchantOrderRateListView[0].MerFee { // 渠道費率與手續費不得高於商戶所設定的
				return 0, errorz.New(response.RATE_SETTING_ERROR)
			}
		}

		// 检查渠道最大内冲金额
		if merchantOrderRateListView[0].MaxInternalCharge < order.OrderAmount {
			return 0, errorz.New(response.CHARGE_AMT_EXCEED)
		}

		// 交易手續費總額 = 訂單金額 / 100 * 費率
		transferHandling := utils.FloatMul(utils.FloatDiv(order.OrderAmount, 100), merchantOrderRateListView[0].MerFee)

		// 計算實際交易金額 = 訂單金額 - 手續費
		transferAmount := utils.FloatSub(order.OrderAmount, transferHandling)

		order.TransferAmount = transferAmount
		order.TransferHandlingFee = transferHandling
		order.ChannelCode = merchantOrderRateListView[0].ChannelCode
		order.ChannelPayTypesCode = merchantOrderRateListView[0].ChannelPayTypesCode
		order.PayTypeCode = merchantOrderRateListView[0].PayTypeCode
		order.Fee = merchantOrderRateListView[0].MerFee
		order.HandlingFee = merchantOrderRateListView[0].MerHandlingFee
		order.BalanceType = constants.DF_BALANCE

		ptBalanceId = merchantOrderRateListView[0].PtBalanceId
	} else {
		if len(req.NcRates) > 0 {
			merMap := make(map[string]string)
			merMap2, errM := l.getAllParentMerchant(merchantCode, merMap)
			if errM != nil {
				return 0, errM
			}

			isRateOk, isProxyOk := l.compareRates(req.NcRates, merMap2)
			if isProxyOk == false {
				return 0, errorz.New(response.MERCHANT_AGENT_NOT_FOUND)
			} else if isRateOk == false {
				return 0, errorz.New(response.SETTING_MERCHANT_RATE_MIN_CHARGE_LOWER_PARENT_ERROR)
			}
		}

		// 交易手續費總額 = 訂單金額 / 100 * 費率
		transferHandling := utils.FloatMul(utils.FloatDiv(order.OrderAmount, 100), req.Rate)

		// 計算實際交易金額 = 訂單金額 - 手續費
		transferAmount := utils.FloatSub(order.OrderAmount, transferHandling)
		merchantOrderRateListView, errR := ordersService.GetMerchantChannelRateByCode(l.svcCtx.MyDB, order.MerchantCode, order.CurrencyCode, req.ChannelCode)
		if errR != nil {
			return 0, errR
		}

		if rateCheck != "0" {
			if req.ChnRate > req.Rate { // 渠道費率與手續費不得高於商戶所設定的
				return 0, errorz.New(response.RATE_SETTING_ERROR)
			}
		}

		// 检查渠道最大内冲金额
		if merchantOrderRateListView.MaxInternalCharge < order.OrderAmount {
			return 0, errorz.New(response.CHARGE_AMT_EXCEED)
		}

		order.TransferAmount = transferAmount
		order.TransferHandlingFee = transferHandling
		order.ChannelCode = merchantOrderRateListView.ChannelCode
		order.ChannelPayTypesCode = merchantOrderRateListView.ChannelPayTypesCode
		order.PayTypeCode = merchantOrderRateListView.PayTypeCode
		order.Fee = req.Rate
		order.HandlingFee = merchantOrderRateListView.MerHandlingFee
		order.BalanceType = constants.XF_BALANCE
		ptBalanceId = merchantOrderRateListView.PtBalanceId
	}

	//编辑订单
	if errC := l.svcCtx.MyDB.Table("tx_orders").Updates(&order).Error; err != nil {
		return 0, errorz.New(response.DATABASE_FAILURE, errC.Error())
	}
	return ptBalanceId, nil
}
