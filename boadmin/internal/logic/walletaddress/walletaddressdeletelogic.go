package walletaddress

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressDeleteLogic {
	return WalletAddressDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressDeleteLogic) WalletAddressDelete(req *types.WalletAddressDeleteRequest) error {
	var walletAddress *types.WalletAddressX

	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Take(&walletAddress, req.ID).Error; err != nil {
		logx.WithContext(l.ctx).Error("删除錢包地址失敗: %s", err.Error())
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if walletAddress.Balance > 0 || walletAddress.Status != "0" {
		logx.WithContext(l.ctx).Error("禁止删除錢包地址: id:%s, Balance:%f, Status:%s ", req.ID, walletAddress.Balance, walletAddress.Status)
		return errorz.New(response.THE_WALLET_ADDRESS_IS_USE_CHANGE_DISABLED)
	}

	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Delete(&req).Error; err != nil {
		logx.WithContext(l.ctx).Error("删除錢包地址失敗: %s", err.Error())
		return errorz.New(response.DELETE_DATABASE_FAILURE, err.Error())
	}

	return nil
}
