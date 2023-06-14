package merchantPtBalance

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceQueryAllLogic {
	return MerchantPtBalanceQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceQueryAllLogic) MerchantPtBalanceQueryAll(req types.MerchantPtBalanceQueryAllRequestX) (resp *types.MerchantPtBalanceQueryAllResponse, err error) {
	var merchantPtBalances []types.MerchantPtBalanceQueryData
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB.Table("mc_merchant_pt_balances mb")

	if len(req.MerchantCode) > 0 {
		db = db.Where("mb.`merchant_code` = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("mb.currency_code = ?", req.CurrencyCode)
	}
	if len(req.Name) > 0 {
		db = db.Where("mb.name = ?", req.Name)
	}

	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	if err = db.
		Select("mb.*").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&merchantPtBalances).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.MerchantPtBalanceQueryAllResponse{
		List:     merchantPtBalances,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
