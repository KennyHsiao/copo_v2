package systemrate

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SystemRateQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSystemRateQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) SystemRateQueryAllLogic {
	return SystemRateQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SystemRateQueryAllLogic) SystemRateQueryAll() (resp *types.SystemRateQueryAllResponse, err error) {
	var systemRate []types.SystemRate

	selectX := "a.code AS currency_code ," +
		"b.id AS ID," +
		"case when b.withdraw_handling_fee IS NULL then 0 else b.withdraw_handling_fee end AS withdraw_handling_fee," +
		"case when b.max_withdraw_charge IS NULL then 0 else b.max_withdraw_charge end AS max_withdraw_charge," +
		"case when b.min_withdraw_charge IS NULL then 0 else b.min_withdraw_charge end AS min_withdraw_charge"

	tx := l.svcCtx.MyDB.Table("bs_currencies a").
		Joins("LEFT JOIN bs_system_rate b on a.code = b.currency_code")

	if err = tx.Select(selectX).Order("sort_order asc").Find(&systemRate).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	resp = &types.SystemRateQueryAllResponse{
		List: systemRate,
	}

	return
}
