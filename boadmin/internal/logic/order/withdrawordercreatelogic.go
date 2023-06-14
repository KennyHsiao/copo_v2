package order

import (
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderCreateLogic {
	return WithdrawOrderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawOrderCreateLogic) WithdrawOrderCreate(req types.MultipleOrderWithdrawCreateRequestX) (resp *types.MultipleOrderCreateResponse, err error) {
	logx.WithContext(l.ctx).Info("页面下发提单： %#v", req)
	//JWT取得登入腳色資訊 用於商戶號
	merchantCode := l.ctx.Value("merchantCode").(string)
	userAccount := l.ctx.Value("account").(string)
	req.List[0].MerchantCode = merchantCode
	req.List[0].UserAccount = userAccount
	var orderWithdrawCreateResp *types.OrderWithdrawCreateResponse
	db := l.svcCtx.MyDB
	if orderWithdrawCreateResp, err = ordersService.WithdrawOrderCreate(db, req.List, l.ctx, l.svcCtx); err != nil {
		return nil, err
	}

	resp = &types.MultipleOrderCreateResponse{
		Index: orderWithdrawCreateResp.Index,
		Errs:  orderWithdrawCreateResp.Errs,
	}
	return
}
