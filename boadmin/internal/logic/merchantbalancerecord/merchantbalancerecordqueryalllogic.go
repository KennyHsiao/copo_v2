package merchantbalancerecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBalanceRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBalanceRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBalanceRecordQueryAllLogic {
	return MerchantBalanceRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBalanceRecordQueryAllLogic) MerchantBalanceRecordQueryAll(req types.MerchantBalanceRecordQueryAllRequestX) (resp *types.MerchantBalanceRecordQueryAllResponse, err error) {
	var merchantBalanceRecords []types.MerchantBalanceRecord
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("mc_merchant_balance_records")

	if req.MerchantBalanceId > 0 {
		//terms = append(terms, fmt.Sprintf("`merchant_balance_id` = '%d'", req.MerchantBalanceId))
		db = db.Where("`merchant_balance_id` = ?", req.MerchantBalanceId)
	}
	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`merchant_code` = '%s'", req.MerchantCode))
		db = db.Where("`merchant_code` = ?", req.MerchantCode)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`currency_code` = '%s'", req.CurrencyCode))
		db = db.Where("`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf("`order_no` = '%s'", req.OrderNo))
		db = db.Where("`order_no` = ?", req.OrderNo)
	}
	if len(req.OrderType) > 0 {
		//terms = append(terms, fmt.Sprintf("`order_type` = '%s'", req.OrderType))
		db = db.Where("`order_type` = ?", req.OrderType)
	}
	if len(req.TransactionType) > 0 {
		//terms = append(terms, fmt.Sprintf("`transaction_type` = '%s'", req.TransactionType))
		db = db.Where("`transaction_type` = ?", req.TransactionType)
	}
	if len(req.BalanceType) > 0 {
		//terms = append(terms, fmt.Sprintf("`balance_type` = '%s'", req.BalanceType))
		db = db.Where("`balance_type` = ?", req.BalanceType)
	}
	if len(req.PayTypeCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`pay_type_code` = '%s'", req.PayTypeCode))
		db = db.Where("`pay_type_code` = ?", req.PayTypeCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("`created_at` >= '%s'", req.StartAt))
		db = db.Where("`created_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("`created_at` < '%s'", endAt))
		db = db.Where("`created_at` < ?", endAt)
	}

	//term := strings.Join(terms, "AND")
	db.Count(&count)
	if err = db.
		Preload("PayTypeData").
		Preload("ChannelData").
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&merchantBalanceRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, record := range merchantBalanceRecords {
		merchantBalanceRecords[i].CreatedAt = utils.ParseTime(record.CreatedAt)
	}

	resp = &types.MerchantBalanceRecordQueryAllResponse{
		List:     merchantBalanceRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
