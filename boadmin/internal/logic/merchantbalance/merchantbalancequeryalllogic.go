package merchantbalance

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceQueryAllLogic {
	return MerchantBalanceQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceQueryAllLogic) MerchantBalanceQueryAll(req types.MerchantBalanceQueryAllRequest) (resp *types.MerchantBalanceQueryAllResponse, err error) {
	var merchantBalances []types.MerchantBalance
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("mc_merchant_balances mb").
		Joins("join mc_merchant_currencies mc on mc.merchant_code = mb.merchant_code and mc.currency_code = mb.currency_code")

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("mb.`merchant_code` like '%%%s%%'", req.MerchantCode))
		db = db.Where("mb.`merchant_code` like ?", "%"+req.MerchantCode+"%")
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("mb.currency_code like '%%%s%%'", req.CurrencyCode))
		db = db.Where("mb.currency_code like ?", "%"+req.CurrencyCode+"%")
	}
	if len(req.BalanceType) > 0 {
		//terms = append(terms, fmt.Sprintf("mb.balance_type = '%s'", req.BalanceType))
		db = db.Where("mb.balance_type = ?", req.BalanceType)
	}
	if len(req.Status) > 0 {
		//terms = append(terms, fmt.Sprintf("mc.status = '%s'", req.Status))
		db = db.Where("mc.status = ?", req.Status)
	}
	//term := strings.Join(terms, " AND ")
	db.Count(&count)
	err = db.
		Select("mb.*").
		Scopes(gormx.Paginate(req)).
		Find(&merchantBalances).Error

	resp = &types.MerchantBalanceQueryAllResponse{
		List:     merchantBalances,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
