package merchant

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryUserBalanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryUserBalanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryUserBalanceLogic {
	return MerchantQueryUserBalanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryUserBalanceLogic) MerchantQueryUserBalance() (resp *types.MerchantQueryListViewResponse, err error) {
	//JWT取得登入账号 和 商戶號
	merchantCode := l.ctx.Value("merchantCode").(string)
	userAccount := l.ctx.Value("account").(string)
	var user types.User
	var currencies []string

	db := l.svcCtx.MyDB.Table("merchant_list_view")

	//取得用戶
	if user, err = l.getUser(userAccount); err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//取得用戶可視幣別
	for _, currency := range user.Currencies {
		currencies = append(currencies, currency.Code)
	}

	var merchants []types.MerchantListView
	//var terms []string

	//terms = append(terms, fmt.Sprintf("code like '%%%s%%'", merchantCode))
	//terms = append(terms, fmt.Sprintf("`currency_code` in ('%s') ", strings.Join(currencies, "','")))
	//terms = append(terms, fmt.Sprintf("`merchant_currencies_status` = '1'"))
	db.Where("code like ?", "%"+merchantCode+"%")
	db.Where("`currency_code` in (?)", currencies)
	db.Where("merchant_currencies_status = '1'")
	//term := strings.Join(terms, " AND ")

	if err = db.Find(&merchants).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantQueryListViewResponse{
		List: merchants,
	}

	return
}

func (l *MerchantQueryUserBalanceLogic) getUser(userAccount string) (user types.User, err error) {
	err = l.svcCtx.MyDB.Table("au_users").
		Where("account = ?", userAccount).
		Preload("Currencies").
		Take(&user).Error
	return
}
