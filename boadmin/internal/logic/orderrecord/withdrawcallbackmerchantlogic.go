package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawCallBackMerchantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawCallBackMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawCallBackMerchantLogic {
	return WithdrawCallBackMerchantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawCallBackMerchantLogic) WithdrawCallBackMerchant(req *types.ProxyPayCallbackMerchantRequest) (err error) {
	var orderX types.OrderX

	logx.WithContext(l.ctx).Errorf("UI下发回调商户, 单号:%s", req.OrderNo)

	if err = l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", req.OrderNo).Take(&orderX).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.DATA_NOT_FOUND, err.Error())
		}
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if orderX.Status == constants.SUCCESS || orderX.Status == constants.FAIL {
		if len(orderX.ChangeType) > 0 && orderX.ChangeType == "1" {
			if err := merchantsService.PostCallbackToMerchant(l.svcCtx.MyDB, &l.ctx, &orderX); err != nil {
				logx.WithContext(l.ctx).Error("UI下发回調商戶錯誤(代付参数):", err)
			}
		} else {
			err := ordersService.WithdrawApiCallBack(l.svcCtx.MyDB, orderX, l.ctx)
			if err != nil {
				logx.WithContext(l.ctx).Error("UI下发回調商戶錯誤:", err)
			}
		}
	} else {
		return errorz.New(response.ORDER_STATUS_WRONG)
	}

	return nil
}
