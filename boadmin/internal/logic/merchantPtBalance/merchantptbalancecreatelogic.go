package merchantPtBalance

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceCreateLogic {
	return MerchantPtBalanceCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceCreateLogic) MerchantPtBalanceCreate(req *types.MerchantPtBalanceCreateRequest) error {

	if err := l.svcCtx.MyDB.Table("mc_merchant_pt_balances").Create(&types.MerchantPtBalanceX{
		MerchantPtBalance: types.MerchantPtBalance{
			MerchantCode: req.MerchantCode,
			CurrencyCode: req.CurrencyCode,
			Name:         req.Name,
			Remark:       req.Remark,
		},
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
