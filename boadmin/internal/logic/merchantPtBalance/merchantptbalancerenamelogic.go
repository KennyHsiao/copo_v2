package merchantPtBalance

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantPtBalanceRenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantPtBalanceRenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantPtBalanceRenameLogic {
	return MerchantPtBalanceRenameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantPtBalanceRenameLogic) MerchantPtBalanceRename(req *types.MerchantPtBalanceRenameRequest) error {
	if err := l.svcCtx.MyDB.Table("mc_merchant_pt_balances").Where("id = ?", req.ID).
		Updates(&types.MerchantPtBalanceX{
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

	return nil
}
