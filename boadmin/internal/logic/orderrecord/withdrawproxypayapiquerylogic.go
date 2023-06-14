package orderrecord

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawProxyPayApiQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawProxyPayApiQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawProxyPayApiQueryLogic {
	return WithdrawProxyPayApiQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawProxyPayApiQueryLogic) WithdrawProxyPayApiQuery(req *types.ProxyPayCallbackMerchantRequest) error {
	// todo: add your logic here and delete this line

	return nil
}
