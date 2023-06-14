package report

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawDetailQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawDetailQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawDetailQueryLogic {
	return WithdrawDetailQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawDetailQueryLogic) WithdrawDetailQuery(req *types.WithdrawCheckBillQueryRequest) (resp *types.WithdrawDetailQueryResponse, err error) {
	var withdrawDetails []types.WithdrawDetail
	//var terms []string
	db := l.svcCtx.MyDB
	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("b.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("b.merchant_code = ?", req.MerchantCode)
	}
	//terms = append(terms, fmt.Sprintf("b.currency_code = '%s'", req.CurrencyCode))
	//terms = append(terms, fmt.Sprintf("b.`created_at` >= '%s'", req.StartAt))
	db = db.Where("b.currency_code = ?", req.CurrencyCode)
	db = db.Where("b.created_at >= ?", req.StartAt)
	endAt := utils.ParseTimeAddOneSecond(req.EndAt)
	//terms = append(terms, fmt.Sprintf("b.`created_at` < '%s'", endAt))
	//terms = append(terms, fmt.Sprintf("b.type = '%s'", constants.ORDER_TYPE_XF))
	//terms = append(terms, fmt.Sprintf("b.is_test != '%s'", constants.IS_TEST_YES))
	//terms = append(terms, fmt.Sprintf("b.reason_type != '%s'", constants.ORDER_REASON_TYPE_RECOVER))
	db = db.Where("b.created_at < ?", endAt)
	db = db.Where("b.type = ?", constants.ORDER_TYPE_XF)
	db = db.Where("b.is_test != ?", constants.IS_TEST_YES)
	db = db.Where("b.reason_type != ?", constants.ORDER_REASON_TYPE_RECOVER)

	//term := strings.Join(terms, " AND ")

	selectX := "a.order_no," +
		"a.order_amount," +
		"b.handling_fee," +
		"b.memo," +
		"b.merchant_code," +
		"b.created_at," +
		"b.trans_at," +
		"c.name AS channel_name"

	if err = db.Select(selectX).Table("tx_order_channels a").
		Joins("join tx_orders b on a.order_no = b.order_no").
		Joins("join ch_channels c on a.channel_code = c.code").
		Find(&withdrawDetails).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//var orders []types.OrderX
	//if err = l.svcCtx.MyDB.Table("tx_orders as b").
	//	Where("b.currency_code = ?", req.CurrencyCode).
	//	Where("b.type = ?", "XF").
	//	Where("b.status = ?", constants.SUCCESS).Find(&orders).Error; err != nil {
	//		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	var totalNum = len(withdrawDetails)
	var totalOrderAmount float64
	var totalHandlingFee float64
	for _, detail := range withdrawDetails {
		totalOrderAmount = utils.FloatAdd(totalOrderAmount, detail.OrderAmount)
		totalHandlingFee = utils.FloatAdd(totalHandlingFee, detail.HandlingFee)
	}

	//for _, order := range orders {
	//	totalHandlingFee = utils.FloatAdd(totalHandlingFee, order.HandlingFee)
	//}

	resp = &types.WithdrawDetailQueryResponse{
		List:             withdrawDetails,
		TotalNum:         totalNum,
		TotalOrderAmount: totalOrderAmount,
		TotalHandlingFee: totalHandlingFee,
	}

	return
}
