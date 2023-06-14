package order

import (
	"com.copo/bo_service/boadmin/internal/service/merchantbalanceservice"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderQueryMerchantChannelFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderQueryMerchantChannelFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderQueryMerchantChannelFeeLogic {
	return OrderQueryMerchantChannelFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderQueryMerchantChannelFeeLogic) OrderQueryMerchantChannelFee(req types.OrderQueryMerchantChannelFeeRequest) (resp *types.OrderQueryMerchantCurrencyAndBanksResponseX, err error) {
	db := l.svcCtx.MyDB
	var merchantOrderRateListViews []*types.MerchantOrderRateListViewX
	//var terms []string

	// 依商户取得对应可用币别
	// JWT取得merchant_code资讯
	var merchantCode string
	if len(req.MerchantCode) > 0 {
		merchantCode = req.MerchantCode
	} else {
		merchantCode = l.ctx.Value("merchantCode").(string)
	}
	// 取得商户渠道设定
	merchantOrderRateListViews, err = ordersService.GetMerchantChannelRate(db, merchantCode, req.CurrencyCode, req.Type)
	if err != nil {
		return nil, err
	}

	var list []types.OrderQueryMerchantCurrencyAndBanks
	for _, view := range merchantOrderRateListViews {
		//取得是哪種錢包
		balanceType, err2 := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, view.ChannelCode, constants.ORDER_TYPE_DF)
		if err2 != nil {
			return nil, err2
		}

		//查詢商戶餘額
		var merchantBalance types.MerchantBalance
		var newTerms []string
		newTerms = append(newTerms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
		newTerms = append(newTerms, fmt.Sprintf("currency_code = '%s'", req.CurrencyCode))
		newTerms = append(newTerms, fmt.Sprintf("balance_type = '%s'", balanceType))
		newTerm := strings.Join(newTerms, " AND ")
		if err = db.Table("mc_merchant_balances").Where(newTerm).Find(&merchantBalance).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		view.Balance = merchantBalance.Balance
		// 查询渠道对应银行资料
		var channelBanks []types.ChannelBankX
		if err = db.Table("ch_channel_banks").Select("bk_banks.bank_no, bk_banks.bank_name, bk_banks.bank_name_en").
			Joins("join bk_banks on bk_banks.bank_no = ch_channel_banks.bank_no").
			Where("ch_channel_banks.channel_code = ? ", view.ChannelCode).
			Where("bk_banks.currency_code = ? ", req.CurrencyCode).
			Find(&channelBanks).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if len(channelBanks) == 0 {
			if err = db.Table("bk_banks").Select("bank_no, bank_name, bank_name_en").
				Where("bk_banks.currency_code = ? ", req.CurrencyCode).
				Find(&channelBanks).Error; err != nil {
				return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		orderQueryMerchantCurrencyANdBanks := types.OrderQueryMerchantCurrencyAndBanks{
			MerchantOrderRateListViewX: view,
			ChannelBanks:               channelBanks,
		}

		list = append(list, orderQueryMerchantCurrencyANdBanks)
	}

	resp = &types.OrderQueryMerchantCurrencyAndBanksResponseX{
		List: list,
	}
	return
}
