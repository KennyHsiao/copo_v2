package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) BatchCheckLogic {
	return BatchCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchCheckLogic) BatchCheck(req types.BatchCheckOrderRequest) (resp *types.BatchCheckOrderRespsonse, err error) {
	ux := model.NewBankBlockAccount(l.svcCtx.MyDB)
	var batchChecks []types.BatchCheck
	var orders = req.List
	var oneCurrencyLimit = req.List[0].CurrencyCode
	var totalPrice float64

	for _, order := range orders {
		//判断是否都是相同币别
		if oneCurrencyLimit != order.CurrencyCode {
			return nil, errorz.New(response.CURRENCY_NOT_THE_SAME)
		}
	}

	// JWT取得登入的 merchantCode资讯
	var merchantCode = l.ctx.Value("merchantCode").(string)
	var handlingFee float64
	var minWithdrawCharge float64
	var maxWithdrawCharge float64
	var isRate string
	var merFee float64
	// 取得商户设定下发手续费
	if req.Type == "XF" {
		var systemRate types.SystemRate
		merchantCurrency := &types.MerchantCurrency{}
		var terms []string

		terms = append(terms, fmt.Sprintf("currency_code = '%s'", oneCurrencyLimit))
		term := strings.Join(terms, " AND ")
		if err = l.svcCtx.MyDB.Table("mc_merchant_currencies").
			Where(term).Where(" merchant_code = ?", merchantCode).Find(&merchantCurrency).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err = l.svcCtx.MyDB.Table("bs_system_rate").Where(term).Take(&systemRate).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		if merchantCurrency.WithdrawHandlingFee <= 0 {
			if systemRate.WithdrawHandlingFee <= 0 {
				return nil, errorz.New(response.MER_WITHDRAW_CHARGE_NOT_SET)
			} else {
				handlingFee = systemRate.WithdrawHandlingFee
			}
		} else {
			handlingFee = merchantCurrency.WithdrawHandlingFee
		}
		minWithdrawCharge = systemRate.MinWithdrawCharge
		maxWithdrawCharge = systemRate.MaxWithdrawCharge
	} else {
		// 取得商戶設定渠道手續費
		var merchantChannelRates []*types.MerchantOrderRateListViewX
		merchantChannelRates, err = ordersService.GetMerchantChannelRate(l.svcCtx.MyDB, merchantCode, oneCurrencyLimit, constants.ORDER_TYPE_DF)
		if err != nil {
			return nil, err
		}
		var merchantChannelRate *types.MerchantOrderRateListViewX
		if len(merchantChannelRates) == 1 {
			if req.TypeSubNo == merchantChannelRates[0].DesignationNo {
				merchantChannelRate = merchantChannelRates[0]
			} else {
				return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
			}
		} else if len(merchantChannelRates) > 1 {
			if len(req.TypeSubNo) > 0 {
				channelRateMap := make(map[string]*types.MerchantOrderRateListViewX)
				for _, view := range merchantChannelRates {
					channelRateMap[view.DesignationNo] = view
				}
				if _, ok := channelRateMap[req.TypeSubNo]; !ok {
					return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
				} else {
					merchantChannelRate = channelRateMap[req.TypeSubNo]
				}
			}
		} else {
			return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
		}
		handlingFee = merchantChannelRate.MerHandlingFee
		isRate = merchantChannelRate.CptIsRate
		merFee = merchantChannelRate.MerFee
	}

	for _, order := range orders {
		// 判斷黑名單
		var isBlock bool
		if isBlock, err = ux.CheckIsBlockAccount(order.MerchantBankAccount); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if !isBlock {
			if req.Type != "XF" {
				// 判斷是否計算費率
				if isRate == "1" {
					// 交易手續費總額 = 訂單金額 / 100 * 費率 + 手續費
					transferHandling := utils.FloatAdd(utils.FloatMul(utils.FloatDiv(order.OrderAmount, 100), merFee), handlingFee)
					order.TransferAmount = utils.FloatAdd(order.OrderAmount, transferHandling)
					totalPrice += order.TransferAmount
				} else {
					order.TransferAmount = utils.FloatAdd(order.OrderAmount, handlingFee)
					totalPrice += order.TransferAmount
				}
			} else {
				order.MinWithdrawCharge = minWithdrawCharge
				order.MaxWithdrawCharge = maxWithdrawCharge
				// 下发仅收手续费
				order.TransferAmount = utils.FloatAdd(order.OrderAmount, handlingFee)
				totalPrice += order.TransferAmount
			}
			order.Valid = true

		} else {
			order.Valid = false
		}
		batchChecks = append(batchChecks, order)
	}

	resp = &types.BatchCheckOrderRespsonse{
		List:        batchChecks,
		TotalPrice:  totalPrice,
		HandlingFee: handlingFee,
		MerFee:      merFee,
		IsRate:      isRate,
	}

	return
}
