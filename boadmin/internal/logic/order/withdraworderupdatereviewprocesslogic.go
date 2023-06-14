package order

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderUpdateReviewProcessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderUpdateReviewProcessLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderUpdateReviewProcessLogic {
	return WithdrawOrderUpdateReviewProcessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawOrderUpdateReviewProcessLogic) WithdrawOrderUpdateReviewProcess(req types.WithdrawOrderUpdateReviewProcssRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		// 取得登入腳色資訊 用於使用者账号
		userAccount := l.ctx.Value("account").(string)
		var order types.Order
		myDB := l.svcCtx.MyDB.Table("tx_orders")
		if err := myDB.Where("order_no = ?", req.OrderNo).Take(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return errorz.New(response.DATA_NOT_FOUND, err.Error())
			}
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if order.Status != "0" {
			return errorz.New(response.ORDER_STATUS_WRONG_CANNOT_PROCESSING, "orderStatus = "+order.Status)
		}

		order.Status = "1" //訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)
		order.UpdatedBy = userAccount
		if err := myDB.Updates(order).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}
