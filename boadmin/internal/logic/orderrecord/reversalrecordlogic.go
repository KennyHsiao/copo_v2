package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReversalRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReversalRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReversalRecordLogic {
	return ReversalRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReversalRecordLogic) ReversalRecord(req types.ReversalRecordRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		//JWT取得使用者資訊
		userAccount := l.ctx.Value("account").(string)
		var order types.OrderX
		//var terms []string

		//terms = append(terms, fmt.Sprintf("`order_no` = '%s'", req.OrderNo))
		//term := strings.Join(terms, "")
		db = db.Where("order_no = ?", req.OrderNo)
		// 查詢訂單資訊
		if err := db.Table("tx_orders").Find(&order).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if order.Status != "20" { // 訂單狀態須為成功才能沖正
			return errorz.New(response.ORDER_STATUS_WRONG_CANNOT_REVERSAL)
		}
		if len(req.Memo) > 0 {
			order.Memo = req.Memo
		}

		if order.BalanceType == "DFB" {
			res, errRpc := l.svcCtx.TransactionRpc.PersonalRebundTransaction_DFB(l.ctx, &transaction.PersonalRebundRequest{
				UserAccount: userAccount,
				OrderNo:     order.OrderNo,
				Memo:        req.Memo,
				Action:      "REVERSAL",
			})
			if errRpc != nil {
				logx.Error("沖正，RPC錯誤", errRpc.Error())
				return errorz.New(response.FAIL, errRpc.Error())
			} else if res.Code != response.API_SUCCESS {
				logx.Errorf("沖正，补回钱包错误: Code:%s, Message:%s", res.Code, res.Message)
				return errorz.New(res.Code, res.Message)
			} else if res.Code == response.API_SUCCESS {
				logx.Infof("沖正，补回DFB钱包成功，单号: %v", res.OrderNo)
			}
		} else if order.BalanceType == "XFB" {
			res, errRpc := l.svcCtx.TransactionRpc.PersonalRebundTransaction_XFB(l.ctx, &transaction.PersonalRebundRequest{
				UserAccount: userAccount,
				OrderNo:     order.OrderNo,
				Memo:        req.Memo,
				Action:      "REVERSAL",
			})
			if errRpc != nil {
				logx.Error("沖正，RPC錯誤", errRpc.Error())
				return errorz.New(response.FAIL, errRpc.Error())
			} else if res.Code != response.API_SUCCESS {
				logx.Errorf("沖正，补回钱包错误: Code:%s, Message:%s", res.Code, res.Message)
				return errorz.New(res.Code, res.Message)
			} else if res.Code == response.API_SUCCESS {
				logx.Infof("沖正，补回XFB钱包成功，单号: %v", res.OrderNo)
			}
		}
		return nil
	})
}
