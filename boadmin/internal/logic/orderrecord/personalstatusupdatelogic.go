package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PersonalStatusUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPersonalStatusUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) PersonalStatusUpdateLogic {
	return PersonalStatusUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PersonalStatusUpdateLogic) PersonalStatusUpdate(req types.PersonalStatusUpdateResponse) error {
	//JWT取得使用者账号
	userAccount := l.ctx.Value("account").(string)
	var order types.OrderX

	if err := l.svcCtx.MyDB.Table("tx_orders").
		Where("`order_no` = ?", req.OrderNo).
		Take(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorz.New(response.DATA_NOT_FOUND, "查无资料，order_no = "+req.OrderNo)
		} else {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if order.Status != constants.TRANSACTION && order.Status != constants.FAIL { // 交易中与失败订单才可更新人工还款
		return errorz.New(response.ORDER_STATUS_WRONG_CANNOT_REPAYMENT)
	}

	if len(req.ChannelOrderNo) > 0 {
		order.ChannelOrderNo = req.ChannelOrderNo
	}
	var jTime types.JsonTime
	if len(req.ChannelTransAt) > 0 {
		timeX, err := jTime.Parse(req.ChannelTransAt)
		if err != nil {
			return errorz.New(response.MERCHANT_COMMISSION_TIME_ERROR)
		}
		order.TransAt = timeX
	} else {
		order.TransAt = jTime.New()
	}

	if len(req.Comment) > 0 {
		order.Memo = req.Comment
	}
	order.PersonProcessStatus = req.PersonProcessStatus //人工处理状态：(0:待處理1:處理中2:成功3:失敗 10:不需处理)
	order.UpdatedBy = userAccount

	var action string
	if req.PersonProcessStatus == "1" {
		action = "PERSON_PROCESSING"
		if err := l.updateOrderStatus(l.svcCtx.MyDB, order, action, userAccount, req.Comment); err != nil {
			return err
		}
	} else if req.PersonProcessStatus == "2" {
		action = "PROCESS_SUCCESS"
		order.Status = "20" //訂單狀態(0:待處理 1:處理中 20:成功 30:失敗 31:凍結)

		if err := l.updateOrderStatus(l.svcCtx.MyDB, order, action, userAccount, req.Comment); err != nil {
			return err
		}
	} else if req.PersonProcessStatus == "3" {

		if err := l.updateOrderStatus(l.svcCtx.MyDB, order, action, userAccount, req.Comment); err != nil {
			return err
		}

		var errRpc error
		var res *transaction.PersonalRebundResponse
		if order.BalanceType == "DFB" {
			res, errRpc = l.svcCtx.TransactionRpc.PersonalRebundTransaction_DFB(l.ctx, &transaction.PersonalRebundRequest{
				UserAccount: userAccount,
				OrderNo:     order.OrderNo,
				Memo:        req.Comment,
			})

			if errRpc != nil {
				logx.Error("人工还款，RPC錯誤", errRpc.Error())
				return errorz.New(response.FAIL, errRpc.Error())
			} else if res.Code != response.API_SUCCESS {
				logx.Errorf("人工还款，补回钱DFB包错误 Code:%s, Message:%s", res.Code, res.Message)
				return errorz.New(res.Code, res.Message)
			} else if res.Code == response.API_SUCCESS {
				logx.Infof("人工还款，补回DFB钱包成功，单号: %v", res.OrderNo)
			}
		} else if order.BalanceType == "XFB" {
			res, errRpc = l.svcCtx.TransactionRpc.PersonalRebundTransaction_DFB(l.ctx, &transaction.PersonalRebundRequest{
				UserAccount: userAccount,
				OrderNo:     order.OrderNo,
				Memo:        req.Comment,
				Action:      "PROCESS_FAIL",
			})

			if errRpc != nil {
				logx.Error("人工还款，RPC錯誤", errRpc.Error())
				return errorz.New(response.FAIL, errRpc.Error())
			} else if res.Code != response.API_SUCCESS {
				logx.Errorf("人工还款，补回钱XFB包错误 Code:%s, Message:%s", res.Code, res.Message)
				return errorz.New(res.Code, res.Message)
			} else if res.Code == response.API_SUCCESS {
				logx.Infof("人工还款，补回XFB钱包成功，单号: %v", res.OrderNo)
			}
		}
	}

	return nil
}

func (l *PersonalStatusUpdateLogic) updateOrderStatus(db *gorm.DB, order types.OrderX, action string, userAccount string, memo string) error {
	return db.Transaction(func(db *gorm.DB) error {

		if err := db.Table("tx_orders").Updates(order).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		//记录订单历程
		orderAction := types.OrderAction{
			OrderNo:     order.OrderNo,
			Action:      action,
			UserAccount: userAccount,
			Comment:     memo,
		}
		if err := model.NewOrderAction(db).CreateOrderAction(&types.OrderActionX{
			OrderAction: orderAction,
		}); err != nil {
			return err
		}
		return nil
	})
}
