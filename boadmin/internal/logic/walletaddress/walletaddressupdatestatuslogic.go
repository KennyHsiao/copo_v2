package walletaddress

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressUpdateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressUpdateStatusLogic {
	return WalletAddressUpdateStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressUpdateStatusLogic) WalletAddressUpdateStatus(req *types.WalletAddressUpdateStatusRequest) error {
	var walletAddressX *types.WalletAddressX
	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Take(&walletAddressX, req.ID).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if walletAddressX.Status == "1" {
		return errorz.New(response.THE_WALLET_ADDRESS_IS_USE_CHANGE_DISABLED)
	}

	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Where("id = ?", req.ID).
		Updates(map[string]interface{}{
			"status": req.Status,
		}).Error; err != nil {
		logx.WithContext(l.ctx).Error("錢包地址 更改失敗失敗: %s", err.Error())
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
