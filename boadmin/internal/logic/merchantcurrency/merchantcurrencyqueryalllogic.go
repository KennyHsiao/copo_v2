package merchantcurrency

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantCurrencyQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantCurrencyQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantCurrencyQueryAllLogic {
	return MerchantCurrencyQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantCurrencyQueryAllLogic) MerchantCurrencyQueryAll(req types.MerchantCurrencyQueryAllRequestX) (resp *types.MerchantCurrencyQueryAllResponse, err error) {
	var merchantCurrencies []types.MerchantCurrency
	var count int64

	db := l.svcCtx.MyDB.Table("mc_merchant_currencies")
	if len(req.MerchantCode) > 0 {
		db = db.Where("`merchant_code` like ?", "%"+req.MerchantCode+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("`currency_code` like ?", "%"+req.CurrencyCode+"%")
	}
	if len(req.Status) > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if len(req.IsDisplayPtBalance) > 0 {
		db = db.Where("is_display_pt_balance = ?", req.IsDisplayPtBalance)
	}

	db.Count(&count)
	if err = db.Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&merchantCurrencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantCurrencyQueryAllResponse{
		List:     merchantCurrencies,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
