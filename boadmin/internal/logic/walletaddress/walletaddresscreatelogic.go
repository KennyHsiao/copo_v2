package walletaddress

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressCreateLogic {
	return WalletAddressCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressCreateLogic) WalletAddressCreate(req *types.WalletAddressCreateRequest) error {
	var walletAddress types.WalletAddressX
	copier.Copy(&walletAddress, &req)
	walletAddress.Status = "2"
	walletAddress.UsageAt = types.JsonTime{}.New()
	walletAddress.ReleaseAt = types.JsonTime{}.New()

	if err := l.svcCtx.MyDB.Table("ch_wallet_address").Create(&walletAddress).Error; err != nil {
		logx.WithContext(l.ctx).Error("創建錢包地址失敗: %s", err.Error())
		if errMySQL, ok := err.(*mysql.MySQLError); ok {
			if errMySQL.Number == 1062 {
				return errorz.New(response.DUPLICATE_WALLET_ADDRESS, err.Error())
			}
		}
		return errorz.New(response.CREATE_FAILURE, err.Error())
	}

	return nil
}
