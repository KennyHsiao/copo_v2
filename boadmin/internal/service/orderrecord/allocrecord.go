package orderrecordService

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"gorm.io/gorm"
)

func AllocRecordQueryAll(db *gorm.DB, req types.AllocRecordQueryAllRequestX) (resp *types.AllocRecordQueryAllResponseX, err error) {
	var allocRecords []types.AllocRecordX
	var count int64
	//var terms []string

	selectX := "tx.*, " +
		"c.name as channel_name "

	tx := db.Table("tx_orders as tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code")

	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.order_no = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf("tx.order_no = '%s'", req.OrderNo))
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.currency_code = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.trans_at > ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at >= '%s'", req.StartAt))
		} else {
			tx = tx.Where("tx.created_at >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at >= '%s'", req.StartAt))
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.trans_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at < '%s'", endAt))
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.created_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at < '%s'", endAt))
		}
	}
	if len(req.Status) > 0 {
		tx = tx.Where("tx.status = ?", req.Status)
		//terms = append(terms, fmt.Sprintf("tx.status = '%s'", req.Status))
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.Type = ?", req.Type)
		//terms = append(terms, fmt.Sprintf("tx.Type = '%s'", req.Type))
	} else {
		tx = tx.Where("tx.Type IN ('BK')")
		//terms = append(terms, fmt.Sprintf("tx.Type in ('BK')"))
	}
	if len(req.Source) > 0 {
		tx = tx.Where("tx.Source = ?", req.Source)
		//terms = append(terms, fmt.Sprintf("tx.Source = '%s'", req.Source))
	}

	//term := strings.Join(terms, " AND ")

	if len(req.ChannelName) > 0 {
		tx = tx.Where("(c.name like ?)", "%"+req.ChannelName+"%").Group("tx.order_no")

		//terms = append(terms, fmt.Sprintf("(c.name like '%%%s%%')", req.ChannelName))
		//term = strings.Join(terms, " AND ")
		//term = term + fmt.Sprintf(" GROUP BY tx.order_no")
	}

	//tx.Where(term)

	if err = tx.Distinct("tx.order_no").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.Distinct("tx.order_no").Select(selectX).
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&allocRecords).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, record := range allocRecords {
		allocRecords[i].CreatedAt = utils.ParseTime(record.CreatedAt)
	}

	resp = &types.AllocRecordQueryAllResponseX{
		List:     allocRecords,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}

func AllocRecordTotalInfo(db *gorm.DB, req types.AllocRecordQueryAllRequestX) (resp *types.AllocRecordTotalInfoResponse, err error) {
	//var terms []string

	selectX := "sum(tx.order_amount) as total_order_amount," +
		"sum(tx.transfer_handling_fee) as total_transfer_handling_fee"

	tx := db.Table("tx_orders tx").
		Joins("LEFT JOIN ch_channels c ON  tx.channel_code = c.code")

	if len(req.OrderNo) > 0 {
		tx = tx.Where("tx.order_no = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf("tx.order_no = '%s'", req.OrderNo))
	}
	if len(req.CurrencyCode) > 0 {
		tx = tx.Where("tx.currency_code = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf("tx.currency_code = '%s'", req.CurrencyCode))
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			tx = tx.Where("tx.trans_at >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at >= '%s'", req.StartAt))
		} else {
			tx = tx.Where("tx.created_at >= ?", req.StartAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at >= '%s'", req.StartAt))
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.trans_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.trans_at < '%s'", endAt))
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			tx = tx.Where("tx.created_at < ?", endAt)
			//terms = append(terms, fmt.Sprintf("tx.created_at < '%s'", endAt))
		}
	}
	if len(req.Status) > 0 {
		tx = tx.Where("tx.status = ?", req.Status)
		//terms = append(terms, fmt.Sprintf("tx.status = '%s'", req.Status))
	}
	if len(req.Type) > 0 {
		tx = tx.Where("tx.type = ?", req.Type)
		//terms = append(terms, fmt.Sprintf("tx.Type = '%s'", req.Type))
	} else {
		tx = tx.Where("tx.type IN ('BK')")
		//terms = append(terms, fmt.Sprintf("tx.Type in ('BK')"))
	}
	if len(req.Source) > 0 {
		tx = tx.Where("tx.Source = ?", req.Source)
		//terms = append(terms, fmt.Sprintf("tx.Source = '%s'", req.Source))
	}

	tx = tx.Where("tx.is_test != ?", constants.IS_TEST_YES)
	//terms = append(terms, fmt.Sprintf("tx.is_test != '%s'", constants.IS_TEST_YES))

	//term := strings.Join(terms, " AND ")

	if len(req.ChannelName) > 0 {
		tx = tx.Where("tx.Source like ?", "%"+req.ChannelName+"%")
		//terms = append(terms, fmt.Sprintf("(c.name like '%%%s%%')", req.ChannelName))
		//term = strings.Join(terms, " AND ")
	}

	if err = tx.Select(selectX).Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
