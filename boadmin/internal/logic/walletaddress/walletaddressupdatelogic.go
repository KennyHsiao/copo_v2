package walletaddress

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressUpdateLogic {
	return WalletAddressUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressUpdateLogic) WalletAddressUpdate(req *types.WalletAddressUpdateRequest) error {
	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Where("id = ?", req.ID).
		Updates(map[string]interface{}{
			"account":     req.Account,
			"address":     req.Address,
			"status":      req.Status,
			"remark":      req.Remark,
			"private_key": req.PrivateKey,
		}).Error; err != nil {
		logx.WithContext(l.ctx).Error("編輯錢包地址失敗: %s", err.Error())
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
