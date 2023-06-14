package merchantPtBalance

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceUpdateLogic {
	return MerchantPtBalanceUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceUpdateLogic) MerchantPtBalanceUpdate(req *types.MerchantPtBalanceUpdateRequest) (err error) {
	if !l.ctx.Value("isAdmin").(bool) {
		// 非管理員禁止使用
		return errorz.New(response.ILLEGAL_REQUEST)
	}

	if isEnable, err := model.NewMerchantCurrency(l.svcCtx.MyDB).IsEnableDisplayPtBalance(req.MerchantCode, req.CurrencyCode); err != nil {
		return errorz.New(response.ILLEGAL_REQUEST)
	} else if !isEnable {
		return errorz.New(response.SUB_WALLET_NOT_ENABLED_THEREFORE_OPERATION_PROHIBITED)
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		newOrderNo := model.GenerateOrderNo("TJ")
		// 1. 取得 商戶子錢包餘額表
		var merchantPtBalance types.MerchantPtBalance

		if err = l.svcCtx.MyDB.Table("mc_merchant_pt_balances").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("merchant_code = ? AND currency_code = ?", req.MerchantCode, req.CurrencyCode).
			Where("name = ?", req.Name).
			Take(&merchantPtBalance).Error; err != nil {
			logx.WithContext(l.ctx).Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		beforeBalance := merchantPtBalance.Balance
		afterBalance := utils.FloatAdd(beforeBalance, req.Amount)
		merchantPtBalance.Balance = afterBalance

		// 2. 變更 子錢包餘額
		if err = db.Table("mc_merchant_pt_balances").Select("balance").Updates(types.MerchantPtBalanceX{
			MerchantPtBalance: merchantPtBalance,
		}).Error; err != nil {
			logx.WithContext(l.ctx).Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 3. 取得 商戶總餘額表
		var merchantBalance types.MerchantBalance
		if err = db.Table("mc_merchant_balances").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("merchant_code = ? AND currency_code = ? AND balance_type = ?", req.MerchantCode, req.CurrencyCode, req.BalanceType).
			Take(&merchantBalance).Error; err != nil {
			logx.WithContext(l.ctx).Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		beforeBalance2 := merchantBalance.Balance
		afterBalance2 := utils.FloatAdd(beforeBalance2, req.Amount)
		merchantBalance.Balance = afterBalance2

		// 4. 變更 商戶總餘額
		if err = db.Table("mc_merchant_balances").Select("balance").Updates(types.MerchantBalanceX{
			MerchantBalance: merchantBalance,
		}).Error; err != nil {
			logx.Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 5. 新增 子錢包餘額紀錄
		merchantPtBalanceRecord := types.MerchantPtBalanceRecord{
			MerchantPtBalanceId: merchantPtBalance.ID,
			MerchantCode:        req.MerchantCode,
			CurrencyCode:        req.CurrencyCode,
			OrderNo:             newOrderNo,
			MerchantOrderNo:     "",
			OrderType:           "TJ",
			ChannelCode:         "",
			PayTypeCode:         "",
			TransactionType:     constants.TRANSACTION_TYPE_ADJUST,
			BeforeBalance:       beforeBalance,
			TransferAmount:      req.Amount,
			AfterBalance:        afterBalance,
			Comment:             req.Comment,
			CreatedBy:           l.ctx.Value("account").(string),
		}

		if err = db.Table("mc_merchant_pt_balance_records").Create(&types.MerchantPtBalanceRecordX{
			MerchantPtBalanceRecord: merchantPtBalanceRecord,
		}).Error; err != nil {
			logx.WithContext(l.ctx).Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 6. 新增 餘額紀錄
		merchantBalanceRecord := types.MerchantBalanceRecord{
			MerchantBalanceId: merchantBalance.ID,
			MerchantCode:      req.MerchantCode,
			CurrencyCode:      req.CurrencyCode,
			OrderNo:           newOrderNo,
			MerchantOrderNo:   "",
			OrderType:         "TJ",
			ChannelCode:       "",
			PayTypeCode:       "",
			TransactionType:   constants.TRANSACTION_TYPE_ADJUST,
			BalanceType:       req.BalanceType,
			BeforeBalance:     beforeBalance2,
			TransferAmount:    req.Amount,
			AfterBalance:      afterBalance2,
			Comment:           req.Comment,
			CreatedBy:         l.ctx.Value("account").(string),
		}

		if err = db.Table("mc_merchant_balance_records").Create(&types.MerchantBalanceRecordX{
			MerchantBalanceRecord: merchantBalanceRecord,
		}).Error; err != nil {
			logx.WithContext(l.ctx).Error(err.Error())
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})
}
