package merchantPtBalance

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceEnableLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceEnableLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceEnableLogic {
	return MerchantPtBalanceEnableLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceEnableLogic) MerchantPtBalanceEnable(req *types.MerchantPtEnableRequest) error {

	logx.WithContext(l.ctx).Infof("MerchantPtBalanceEnable: %+v", req)

	if !l.ctx.Value("isAdmin").(bool) {
		// 非管理員禁止使用
		return errorz.New(response.ILLEGAL_REQUEST)
	}

	var merchant types.Merchant

	// 商户禁用状态才能启用子钱包功能
	if err := l.svcCtx.MyDB.Table("mc_merchants").
		Where("code = ? AND status = ?", req.MerchantCode, "0").
		Take(&merchant).Error; err != nil {
		return errorz.New("商户禁用状态才能启用子钱包功能", err.Error())
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		var merchantCurrency types.MerchantCurrency
		var ptBalanceMap = map[string]int64{}
		totalBalance := 0.0
		totalPtBalance := 0.0

		// 1. 取得商户币别
		if err = db.Table("mc_merchant_currencies").
			Where("merchant_code = ? AND currency_code = ?", req.MerchantCode, req.CurrencyCode).
			Take(&merchantCurrency).Error; err != nil {
			return errorz.New(response.ILLEGAL_REQUEST, err.Error())
		} else if merchantCurrency.IsDisplayPtBalance == "1" {
			// 已启用报错
			return errorz.New(response.SUB_WALLET_ENABLED_THEREFORE_OPERATION_PROHIBITED)
		}

		// 2. 新增子钱包
		for _, ptBalance := range req.MerchantPtBalances {
			totalPtBalance += ptBalance.Balance
			ptBalance.MerchantCode = req.MerchantCode
			ptBalance.CurrencyCode = req.CurrencyCode
			merchantPtBalanceX := types.MerchantPtBalanceX{
				MerchantPtBalance: ptBalance,
			}
			if err = db.Table("mc_merchant_pt_balances").Create(&merchantPtBalanceX).Error; err != nil {
				logx.WithContext(l.ctx).Infof("MerchantPtBalanceEnable error: %s", err.Error())
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			ptBalanceMap[merchantPtBalanceX.Name] = merchantPtBalanceX.ID
		}

		// 3. 取得商户大钱包总金额
		if err = db.Select("sum(balance) as balance").Table("mc_merchant_balances").
			Where("merchant_code = ? AND currency_code = ? AND balance_type in ('DFB','XFB')", req.MerchantCode, req.CurrencyCode).
			Find(&totalBalance).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 4.比对小钱包总额是否等于大钱包总额
		if totalBalance != totalPtBalance {
			return errorz.New("大钱包与小钱包金额不符")
		}

		// 5.update 商户渠道费率使用的子钱包
		for _, ptBalance := range req.MerchantPtChannels {
			if err = db.Table("mc_merchant_channel_rate").Where("id = ?", ptBalance.ID).
				Updates(map[string]interface{}{"merchant_pt_balance_id": ptBalanceMap[ptBalance.PtBalanceName]}).Error; err != nil {
				logx.WithContext(l.ctx).Infof("MerchantPtBalanceEnable error: %s", err.Error())
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		// 6.确认是否有费率没设置子钱包
		isHas := false
		if err = db.Table("mc_merchant_channel_rate").Select("count(*) > 0").
			Where("merchant_code = ? AND currency_code = ? AND (merchant_pt_balance_id is null or merchant_pt_balance_id = '' )",
				req.MerchantCode, req.CurrencyCode).
			Find(&isHas).Error; err != nil {
			logx.WithContext(l.ctx).Infof("MerchantPtBalanceEnable error: %s", err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		} else if isHas {
			return errorz.New("有未配置子钱包的费率")
		}

		// 7.update mc_merchant_currencies
		merchantCurrency.IsDisplayPtBalance = "1"
		if err = db.Table("mc_merchant_currencies").Where("id = ?", merchantCurrency.ID).
			Updates(merchantCurrency).Error; err != nil {
			logx.WithContext(l.ctx).Infof("MerchantPtBalanceEnable error: %s", err.Error())
			return errorz.New(response.DATABASE_FAILURE)
		}

		return
	})
}
