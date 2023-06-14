package admin_user

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserCurrenciesQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserCurrenciesQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserCurrenciesQueryLogic {
	return UserCurrenciesQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserCurrenciesQueryLogic) UserCurrenciesQuery(req types.UserCurrenciesQueryRequest) (resp *types.UserCurrenciesQueryResponse, err error) {
	var userCurrencies []types.UserCurrencyQuery

	if err = l.svcCtx.MyDB.
		Select("auc.user_account, mmc.currency_code, mmc.sort_order").
		Table("au_user_merchants aum").
		Joins("JOIN mc_merchant_currencies mmc ON aum.merchant_code = mmc.merchant_code").
		Joins("JOIN au_user_currencies auc ON  auc.user_account = aum.user_account and auc.currency_code = mmc.currency_code").
		Where("aum.user_account = ?", req.UserAccount).
		Order("sort_order").
		Find(&userCurrencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.UserCurrenciesQueryResponse{
		List: userCurrencies,
	}

	return
}
