package order

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderFeeLogic {
	return WithdrawOrderFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawOrderFeeLogic) WithdrawOrderFee(req types.WithdrawOrderFeeRequest) (resp *types.WithdrawOrderFeeResponse, err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	var res types.SystemRate
	var merchantCurrency types.MerchantCurrency

	// 商户下发手续费
	if err = l.svcCtx.MyDB.Table("mc_merchant_currencies").
		Where("currency_code = ?", req.CurrencyCode).
		Where(" merchant_code = ?", merchantCode).
		Find(&merchantCurrency).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	// 下发设定
	myDB := l.svcCtx.MyDB.Table("bs_system_rate").Where("currency_code = ?", req.CurrencyCode)

	if err = myDB.Take(&res).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errorz.New(response.SYSTEM_RATE_NOT_SET, err.Error())
	} else if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	var handlingFee float64
	if merchantCurrency.WithdrawHandlingFee <= 0 {
		if res.WithdrawHandlingFee <= 0 {
			return nil, errorz.New(response.MER_WITHDRAW_CHARGE_NOT_SET)
		} else {
			handlingFee = res.WithdrawHandlingFee
		}
	} else {
		handlingFee = merchantCurrency.WithdrawHandlingFee
	}

	// 取得商户余额
	var merchantBalance types.MerchantBalance
	var merchantPtBalances []types.MerchantPtBalance

	if err = l.svcCtx.MyDB.Table("mc_merchant_balances").
		Where("merchant_code =?", merchantCode).
		Where("currency_code = ?", req.CurrencyCode).
		Where("balance_type = 'XFB'").
		Take(&merchantBalance).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = l.svcCtx.MyDB.Table("mc_merchant_currencies").
		Where("merchant_code = ?", merchantCode).
		Where("currency_code = ?", req.CurrencyCode).
		Find(&merchantCurrency).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	isDisplayPtBalance := merchantCurrency.IsDisplayPtBalance

	if isDisplayPtBalance == "1" {
		if err = l.svcCtx.MyDB.Table("mc_merchant_pt_balances").
			Where("merchant_code = ?", merchantCode).
			Where("currency_code = ?", req.CurrencyCode).
			Find(&merchantPtBalances).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	resp = &types.WithdrawOrderFeeResponse{
		HandlerFee:        handlingFee,
		MinWithdrawCharge: res.MinWithdrawCharge,
		MaxWithdrawCharge: res.MaxWithdrawCharge,
		Balance:           merchantBalance.Balance,
		PtBalance:         merchantPtBalances,
	}
	return
}

type SelectX struct {
	MerchantCode        string  `json:"merchant_code"`
	WithdrawHandlingFee float64 `json:"withdraw_handling_fee"`
	MinWithdrawCharge   float64 `json:"min_withdraw_charge"`
	MaxWithdrawCharge   float64 `json:"max_withdraw_charge"`
}
