package walletaddress

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressQueryLogic {
	return WalletAddressQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressQueryLogic) WalletAddressQuery(req *types.WalletAddressQueryRequest) (resp *types.WalletAddressX, err error) {
	if err = l.svcCtx.MyDB.Table("ch_wallet_address").Take(&resp, req.ID).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
