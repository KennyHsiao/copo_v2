package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayCallBackMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayCallBackMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayCallBackMerchantLogic {
	return ProxyPayCallBackMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayCallBackMerchantLogic) ProxyPayCallBackMerchant(req *types.ProxyPayCallbackMerchantRequest) (err error) {
	var orderX types.OrderX

	logx.WithContext(l.ctx).Errorf("UI代付回调商户, 单号:%s", req.OrderNo)

	if err = l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", req.OrderNo).Take(&orderX).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL {
		if err = merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, &orderX); err != nil {
			logx.Error("UI代付回調商戶錯誤:", err)
			return errorz.New(response.MERCHANT_CALLBACK_ERROR, err.Error())
		}
	} else {
		return errorz.New(response.ORDER_STATUS_WRONG)
	}

	return nil
}
