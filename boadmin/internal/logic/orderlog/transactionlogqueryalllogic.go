package orderlog

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransactionLogQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransactionLogQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) TransactionLogQueryAllLogic {
	return TransactionLogQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransactionLogQueryAllLogic) TransactionLogQueryAll(req *types.OrderLogQueryAllRequestX) (resp *types.OrderLogQueryAllResponseX, err error) {
	var txLogs []types.TxLog
	var count int64
	db := l.svcCtx.MyDB

	// 若给商户订单号 自动带入平台订单号 (平台订单号反之)
	if len(req.MerchantOrderNo) > 0 {
		orderNo := ""
		//l.svcCtx.MyDB.Distinct("order_no").
		//	Table("tx_orders").
		//	Where("merchant_order_no = ?", req.MerchantOrderNo).
		//	Find(&orderNo)
		if orderNo != "" {
			db = db.Where("merchant_order_no = ? or order_no = ?", req.MerchantOrderNo, orderNo)
		} else {
			db = db.Where("merchant_order_no = ?", req.MerchantOrderNo)
		}

	} else if len(req.OrderNo) > 0 {
		merchantOrderNo := ""
		l.svcCtx.MyDB.Distinct("merchant_order_no").
			Table("tx_orders").
			Where("order_no = ?", req.OrderNo).
			Find(&merchantOrderNo)
		if merchantOrderNo != "" {
			db = db.Where("merchant_order_no = ? or order_no = ?", merchantOrderNo, req.OrderNo)
		} else {
			db = db.Where("order_no = ?", req.OrderNo)
		}
	}

	if len(req.MerchantCode) > 0 {
		db = db.Where("merchant_code = ?", req.MerchantCode)
	}

	if len(req.StartAt) > 0 {
		db = db.Where("`created_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		db = db.Where("`created_at` < ?", endAt)
	}
	if len(req.LogSource) > 0 {
		db = db.Where("log_source = ?", req.LogSource)
	}
	if len(req.LogType) > 0 {
		db = db.Where("log_type IN ?", req.LogType)
	}

	db.Table("tx_log").Count(&count)

	if err = db.Table("tx_log").
		Scopes(gormx.Paginate(req.OrderLogQueryAllRequest)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&txLogs).Error; err != nil {
		return
	}

	return &types.OrderLogQueryAllResponseX{
		List:     txLogs,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}, nil
}
