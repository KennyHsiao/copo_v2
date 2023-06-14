package merchant

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryUserPtBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryUserPtBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryUserPtBalanceLogic {
	return MerchantQueryUserPtBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryUserPtBalanceLogic) MerchantQueryUserPtBalance() (resp *types.MerchantPtBalanceListResponse, err error) {
	//JWT取得登入账号 和 商戶號
	merchantCode := l.ctx.Value("merchantCode").(string)
	userAccount := l.ctx.Value("account").(string)
	var user types.User
	var currencies []string
	db := l.svcCtx.MyDB.Table("mc_merchant_pt_balances mb")
	//取得用戶
	if user, err = l.getUser(userAccount); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//取得用戶可視幣別
	for _, currency := range user.Currencies {
		currencies = append(currencies, currency.Code)
	}

	var merchantPtBalances []types.MerchantPtBalance
	//var terms []string

	db = db.Where("mb.`merchant_code` = ?", merchantCode)
	db = db.Where("mb.`currency_code` in (?)", currencies)

	if err = db.
		Select("mb.*").
		Order("mb.created_at asc").
		Find(&merchantPtBalances).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.MerchantPtBalanceListResponse{
		List: merchantPtBalances,
	}

	return
}

func (l *MerchantQueryUserPtBalanceLogic) getUser(userAccount string) (user types.User, err error) {
	err = l.svcCtx.MyDB.Table("au_users").
		Where("account = ?", userAccount).
		Preload("Currencies").
		Take(&user).Error
	return
}
