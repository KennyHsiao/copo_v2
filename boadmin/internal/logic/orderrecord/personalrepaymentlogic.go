package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type PersonalRepaymentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPersonalRepaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) PersonalRepaymentLogic {
	return PersonalRepaymentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PersonalRepaymentLogic) PersonalRepayment(req types.PersonalRepaymentRequestX) (resp *types.PersonalRepaymentResponseX, err error) {
	var personalRepayments []types.PersonalRepaymentX
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB

	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.order_no = '%s'", req.OrderNo))
		db = db.Where("tx.order_no = ?", req.OrderNo)
	}
	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf("c.name like '%%%s%%'", req.ChannelName))
		db = db.Where("c.name like ?", "%"+req.ChannelName+"%")
	}
	if len(req.Source) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.source = '%s'", req.Source))
		db = db.Where("tx.source = ?", req.Source)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.created_at >= '%s'", req.StartAt))
		db = db.Where("tx.created_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("tx.created_at < '%s'", endAt))
		db = db.Where("tx.created_at < ?", endAt)
	}

	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("tx.currency_code = ?", req.CurrencyCode)
	}

	//terms = append(terms, fmt.Sprintf("tx.person_process_status IN ('%s','%s')", "0", "1")) // 人工处理状态：(0:待處理1:處理中2:成功3:失敗 10:不需处理)
	//terms = append(terms, fmt.Sprintf("tx.type = 'DF'"))
	//terms = append(terms, fmt.Sprintf("tx.status IN ('%s', '%s')", constants.TRANSACTION, constants.FAIL))
	db = db.Where("tx.person_process_status IN (?,?)", "0", "1")
	db = db.Where("tx.type = ?", "DF")
	db = db.Where("tx.status IN (?,?)", constants.TRANSACTION, constants.FAIL)
	//term := strings.Join(terms, " AND ")

	selectX := "tx.*, " +
		"c.name as channel_name "

	db2 := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code")

	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db2.Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&personalRepayments).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.PersonalRepaymentResponseX{
		List:     personalRepayments,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
