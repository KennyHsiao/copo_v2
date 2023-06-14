package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/downloadReportService"
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/excelizeutil"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"encoding/json"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/xuri/excelize/v2"
	"strconv"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductRecordExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeductRecordExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) DeductRecordExcelLogic {
	return DeductRecordExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeductRecordExcelLogic) DeductRecordExcel(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, excelName string, err error) {

	// 設置i18n
	utils.SetI18n(req.Language)

	merchantCode := l.ctx.Value("merchantCode").(string)
	userId := l.ctx.Value("userId").(json.Number)
	userIdint, _ := userId.Int64()
	isAdmin := l.ctx.Value("isAdmin").(bool)
	requestBytes, _ := json.Marshal(req)

	createDownloadTask := &types.CreateDownloadTask{
		Prefix:       "",
		Infix:        "",
		Suffix:       "",
		IsAdmin:      isAdmin,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		CurrencyCode: req.CurrencyCode,
		MerchantCode: merchantCode,
		UserId:       userIdint,
		ReqParam:     string(requestBytes),
		Type:         constants.RECEIPT_RECORD,
	}

	if req.DateType == "1" {
		createDownloadTask.Infix = "提单时间 : "
	} else if req.DateType == "2" {
		createDownloadTask.Infix = "交易时间 : "
	}

	if req.Type == constants.ORDER_TYPE_DF {
		createDownloadTask.Prefix = "代付纪录数据"
		createDownloadTask.Type = constants.PROXYPAY_RECORD
		downReportID, fileName, err2 := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
		if err2 != nil {
			return nil, "", err2
		}
		go func() {
			// 代付紀錄
			xlsx, err := l.newProxyPayRecordExcelExport(req)
			if err != nil {
				logx.WithContext(l.ctx).Error("newProxyPayRecordExcelExport Err: " + err.Error())
			}
			downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
		}()

	} else if req.Type == constants.ORDER_TYPE_XF && req.Status == constants.SUCCESS {
		// 下發明細
		if xlsx, err = l.withdrawDetailExcelExport(req); err != nil {
			return nil, "", errorz.New(response.GENERAL_EXCEPTION)
		}
		excelName = "IssueDetailRecord" + time.Now().Format("20060102150405") + ".xlsx"
	} else if req.Type == constants.ORDER_TYPE_XF {
		createDownloadTask.Prefix = "下发纪录数据"
		createDownloadTask.Type = constants.WITHDRAW_RECORD
		downReportID, fileName, err2 := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
		if err2 != nil {
			return nil, "", err2
		}
		go func() {
			// 下發紀錄
			xlsx, err = l.newWithdrawExcelExport(req)
			if err != nil {
				logx.WithContext(l.ctx).Error("newWithdrawExcelExport Err: " + err.Error())
			}
			downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
		}()
	}

	return
}

func (l *DeductRecordExcelLogic) proxyPayRecordExcelExport(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {

	sheetName := i18n.Sprintf("Pay on behalf of record")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.DeductRecordQueryAllResponseX
	if resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, true, nil); err != nil {
		return
	}

	// 建立excel
	xlsx = excelize.NewFile()

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 設置最候一行 Style
	bottomStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 建立分頁
	xlsx.SetSheetName("Sheet1", sheetName)

	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant order number"),
		i18n.Sprintf("Payment channel"), i18n.Sprintf("Merchant ID"),
		i18n.Sprintf("Account Bank"), i18n.Sprintf("Account opening province"),
		i18n.Sprintf("Account opening city"), i18n.Sprintf("Account name"),
		i18n.Sprintf("Bank card number"), i18n.Sprintf("Withdrawal method"),
		i18n.Sprintf("Withdrawal amount"), i18n.Sprintf("Actual handling fee"), i18n.Sprintf("Handling fee"),
		i18n.Sprintf("Order status"), i18n.Sprintf("Remarks"),
		i18n.Sprintf("Bill of Lading Time"), i18n.Sprintf("Transaction time"),
		i18n.Sprintf("Order source"), i18n.Sprintf("Currency"),
	}

	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	var jsonTime types.JsonTime

	for i, item := range resp.List {
		createdAt, _ := jsonTime.Parse(item.CreatedAt)
		var orderNo = item.OrderNo
		var memo = item.Memo
		if item.IsTest == "1" {
			orderNo = item.OrderNo + "(测)"
			memo = i18n.Sprintf("TRANSFER_TEST")
		}
		row := []interface{}{
			orderNo, item.MerchantOrderNo,
			item.ChannelName, item.MerchantCode,
			item.MerchantBankName, item.MerchantBankProvince,
			item.MerchantBankCity, item.MerchantAccountName,
			item.MerchantBankAccount, excelizeutil.GetTxOrderTypeName(item.Type),
			item.OrderAmount, item.TransferHandlingFee, item.HandlingFee,
			excelizeutil.GetTxOrderStatusName(item.Status), memo,
			createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			excelizeutil.GetTxOrderSourceName(item.Source), item.CurrencyCode,
		}

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	totalAmount := xlsx.SetCellFormula(sheetName, "K"+strconv.Itoa(len(resp.List)+2), "=SUM(K2:"+"K"+strconv.Itoa(len(resp.List)+1)+")")
	totalHandlingFee := xlsx.SetCellFormula(sheetName, "L"+strconv.Itoa(len(resp.List)+2), "=SUM(L2:"+"L"+strconv.Itoa(len(resp.List)+1)+")")
	row := []interface{}{
		i18n.Sprintf("count"), "",
		"", "",
		"", "",
		"", "",
		"", "",
		totalAmount, totalHandlingFee,
		"", "",
		"", "",
		"", "",
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, len(resp.List)+2, rowHeight)
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &row); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)

	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "R1", headerStyle)

	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "R"+footerRow, bottomStyle)

	return
}

func (l *DeductRecordExcelLogic) withdrawExcelExport(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	sheetName := i18n.Sprintf("issued record")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.DeductRecordQueryAllResponseX
	if resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, true, nil); err != nil {
		return
	}

	// 建立excel
	xlsx = excelize.NewFile()

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 設置最候一行 Style
	bottomStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 建立分頁
	xlsx.SetSheetName("Sheet1", sheetName)

	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant ID"),
		i18n.Sprintf("Account opening province"), i18n.Sprintf("Account opening city"),
		i18n.Sprintf("Account name"), i18n.Sprintf("Bank card number"),
		i18n.Sprintf("Withdrawal amount"), i18n.Sprintf("Handling fee"),
		i18n.Sprintf("Order status"),
		i18n.Sprintf("Remarks"), i18n.Sprintf("Bill of Lading Time"),
		i18n.Sprintf("Transaction time"), i18n.Sprintf("Update staff"),
	}

	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	var jsonTime types.JsonTime

	for i, item := range resp.List {
		createdAt, _ := jsonTime.Parse(item.CreatedAt)
		row := []interface{}{
			item.OrderNo, item.MerchantCode,
			item.MerchantBankName, item.MerchantBankProvince,
			item.MerchantAccountName, item.MerchantBankAccount,
			item.OrderAmount, item.TransferHandlingFee,
			excelizeutil.GetTxOrderStatusName(item.Status),
			item.Memo, createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.ReviewedBy,
		}

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	totalAmount := xlsx.SetCellFormula(sheetName, "G"+strconv.Itoa(len(resp.List)+2), "=SUM(G2:"+"G"+strconv.Itoa(len(resp.List)+1)+")")
	totalHandlingFee := xlsx.SetCellFormula(sheetName, "H"+strconv.Itoa(len(resp.List)+2), "=SUM(H2:"+"H"+strconv.Itoa(len(resp.List)+1)+")")
	row := []interface{}{
		i18n.Sprintf("count"), "",
		"", "",
		"", totalAmount,
		totalHandlingFee,
		"", "",
		"", "",
		"", "",
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, len(resp.List)+2, rowHeight)
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &row); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)

	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "M1", headerStyle)

	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "M"+footerRow, bottomStyle)

	return
}

func (l *DeductRecordExcelLogic) withdrawDetailExcelExport(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	sheetName := i18n.Sprintf("Issued detail")
	var rowHeight float64 = 17

	// 取得资料
	var IssueDetailRecord []types.IssueDetailRecord
	db := l.svcCtx.MyDB

	selectX := "a.order_no," +
		"a.order_amount," +
		"b.memo," +
		"b.merchant_code," +
		"b.created_at," +
		"b.trans_at," +
		"c.name"

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, map[string]interface{}{"b.`merchant_code`": req.MerchantCode})
		db = db.Where("b.merchant_code = ?", req.MerchantCode)
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf("a.`order_no` = '%s'", req.OrderNo))
		db = db.Where("a.`order_no` = ?", req.OrderNo)
	}
	if len(req.MerchantOrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf("b.`merchant_order_no` = '%s'", req.MerchantOrderNo))
		db = db.Where("b.`merchant_order_no` = ?", req.MerchantOrderNo)
	}
	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("b.`currency_code` = '%s'", req.CurrencyCode))
		db = db.Where("b.`currency_code` = ?", req.CurrencyCode)
	}
	if len(req.StartAt) > 0 {
		if req.DateType == "2" {
			//terms = append(terms, fmt.Sprintf("b.`trans_at` >= '%s'", req.StartAt))
			db = db.Where("b.`trans_at` >= ?", req.StartAt)
		} else {
			//terms = append(terms, fmt.Sprintf("b.`created_at` >= '%s'", req.StartAt))
			db = db.Where("b.`created_at` >= ?", req.StartAt)
		}
	}
	if len(req.EndAt) > 0 {
		if req.DateType == "2" {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			//terms = append(terms, fmt.Sprintf("b.`trans_at` < '%s'", endAt))
			db = db.Where("b.`trans_at` < ?", endAt)
		} else {
			endAt := utils.ParseTimeAddOneSecond(req.EndAt)
			//terms = append(terms, fmt.Sprintf("b.`created_at` < '%s'", endAt))
			db = db.Where("b.`created_at` < ?", endAt)
		}
	}
	if len(req.Status) > 0 {
		//terms = append(terms, fmt.Sprintf("b.`status` = '%s'", req.Status))
		db = db.Where("b.`status` = ?", req.Status)
	}
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("b.`type` = '%s'", req.Type))
		db = db.Where("b.`type` = ?", req.Type)
	}

	if len(req.ChannelName) > 0 {
		//terms = append(terms, fmt.Sprintf("c.name like '%%%s%%'", req.ChannelName))
		db = db.Where("c.name like ?", "%"+req.ChannelName+"%")
	}

	//term := strings.Join(terms, " AND ")

	if err = db.Select(selectX).Table("tx_order_channels a").
		Joins("join tx_orders b on a.order_no = b.order_no").
		Joins("join ch_channels c on a.channel_code = c.code").
		Find(&IssueDetailRecord).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	// 建立excel
	xlsx = excelize.NewFile()

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 建立分頁
	xlsx.SetSheetName("Sheet1", sheetName)

	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant ID"),
		i18n.Sprintf("Channel Name"), i18n.Sprintf("Withdrawal amount"),
		i18n.Sprintf("Remarks"), i18n.Sprintf("Bill of Lading Time"),
		i18n.Sprintf("Transaction time"),
	}

	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	var jsonTime types.JsonTime
	var totalOrderAmount float64
	for i, item := range IssueDetailRecord {
		createdAt := utils.ParseTime(item.CreatedAt)
		transAt := utils.ParseTime(item.TransAt)
		newCreatedAt, _ := jsonTime.Parse(createdAt)
		newTransAt, _ := jsonTime.Parse(transAt)

		row := []interface{}{
			item.OrderNo, item.MerchantCode,
			item.Name, item.OrderAmount,
			item.Memo, newCreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			newTransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
		}

		totalOrderAmount = utils.FloatAdd(totalOrderAmount, item.OrderAmount)
		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 設置總金額
	totalRow := []interface{}{
		"加总", "-",
		"-", totalOrderAmount,
		"-", "-", "-",
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(IssueDetailRecord)+2), &totalRow); err != nil {
		return
	}

	// 設置標題 style
	xlsx.SetRowStyle(sheetName, 1, 1, headerStyle)
	// 設置頁腳 style
	footerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})
	footerRow := strconv.Itoa(len(IssueDetailRecord) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "G"+footerRow, footerStyle)

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)

	return
}

func (l *DeductRecordExcelLogic) newProxyPayRecordExcelExport(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	sheetName := i18n.Sprintf("Pay on behalf of record")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.DeductRecordQueryAllResponseX
	if resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, true, nil); err != nil {
		return
	}

	// 建立excel
	xlsx = excelize.NewFile()

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 設置最候一行 Style
	bottomStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 建立分頁
	xlsx.SetSheetName("Sheet1", sheetName)
	w, err := xlsx.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}
	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant order number"),
		i18n.Sprintf("Payment channel"), i18n.Sprintf("Merchant ID"),
		i18n.Sprintf("Account Bank"), i18n.Sprintf("Account opening province"),
		i18n.Sprintf("Account opening city"), i18n.Sprintf("Account name"),
		i18n.Sprintf("Bank card number"), i18n.Sprintf("Withdrawal method"),
		i18n.Sprintf("Withdrawal amount"), i18n.Sprintf("Actual handling fee"), i18n.Sprintf("Handling fee"),
		i18n.Sprintf("Order status"), i18n.Sprintf("Remarks"),
		i18n.Sprintf("Bill of Lading Time"), i18n.Sprintf("Transaction time"),
		i18n.Sprintf("Order source"), i18n.Sprintf("Currency"),
	}

	w.SetRow("A1", header,
		excelize.RowOpts{StyleID: headerStyle, Height: rowHeight, Hidden: false})

	// 迴圈建置資料
	var jsonTime types.JsonTime
	var totalAmount, totalHandlingFee float64

	for i, item := range resp.List {
		createdAt, _ := jsonTime.Parse(item.CreatedAt)
		var orderNo = item.OrderNo
		var memo = item.Memo
		if item.IsTest == "1" {
			orderNo = item.OrderNo + "(测)"
			memo = i18n.Sprintf("TRANSFER_TEST")
		}
		row := []interface{}{
			orderNo, item.MerchantOrderNo,
			item.ChannelName, item.MerchantCode,
			item.MerchantBankName, item.MerchantBankProvince,
			item.MerchantBankCity, item.MerchantAccountName,
			item.MerchantBankAccount, excelizeutil.GetTxOrderTypeName(item.Type),
			item.OrderAmount, item.TransferHandlingFee, item.HandlingFee,
			excelizeutil.GetTxOrderStatusName(item.Status), memo,
			createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			excelizeutil.GetTxOrderSourceName(item.Source), item.CurrencyCode,
		}
		totalAmount = utils.FloatAdd(totalAmount, item.OrderAmount)
		totalHandlingFee = utils.FloatAdd(totalHandlingFee, item.TransferHandlingFee)
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		w.SetRow(cell, row)
	}
	//totalAmount := xlsx.SetCellFormula(sheetName, "K"+strconv.Itoa(len(resp.List)+2), "=SUM(K2:"+"K"+strconv.Itoa(len(resp.List)+1)+")")
	//totalHandlingFee := xlsx.SetCellFormula(sheetName, "L"+strconv.Itoa(len(resp.List)+2), "=SUM(L2:"+"L"+strconv.Itoa(len(resp.List)+1)+")")
	row := []interface{}{
		excelize.Cell{StyleID: bottomStyle, Value: i18n.Sprintf("count")}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: totalAmount}, excelize.Cell{StyleID: bottomStyle, Value: totalHandlingFee},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""},
	}
	footerRow := strconv.Itoa(len(resp.List) + 2)
	w.SetRow("A"+footerRow, row)
	w.Flush()

	return
}

func (l *DeductRecordExcelLogic) newWithdrawExcelExport(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	sheetName := i18n.Sprintf("issued record")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.DeductRecordQueryAllResponseX
	if resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, true, nil); err != nil {
		return
	}

	// 建立excel
	xlsx = excelize.NewFile()

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 設置最候一行 Style
	bottomStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 建立分頁
	xlsx.SetSheetName("Sheet1", sheetName)
	w, err := xlsx.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}

	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant ID"),
		i18n.Sprintf("Account opening province"), i18n.Sprintf("Account opening city"),
		i18n.Sprintf("Account name"), i18n.Sprintf("Bank card number"),
		i18n.Sprintf("Withdrawal amount"), i18n.Sprintf("Handling fee"),
		i18n.Sprintf("Order status"),
		i18n.Sprintf("Remarks"), i18n.Sprintf("Bill of Lading Time"),
		i18n.Sprintf("Transaction time"), i18n.Sprintf("Update staff"),
	}

	w.SetRow("A1", header,
		excelize.RowOpts{StyleID: headerStyle, Height: rowHeight, Hidden: false})

	// 迴圈建置資料
	var jsonTime types.JsonTime
	var totalAmount, totalHandlingFee float64
	for i, item := range resp.List {
		createdAt, _ := jsonTime.Parse(item.CreatedAt)
		row := []interface{}{
			item.OrderNo, item.MerchantCode,
			item.MerchantBankName, item.MerchantBankProvince,
			item.MerchantAccountName, item.MerchantBankAccount,
			item.OrderAmount, item.TransferHandlingFee,
			excelizeutil.GetTxOrderStatusName(item.Status),
			item.Memo, createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.ReviewedBy,
		}
		totalAmount = utils.FloatAdd(totalAmount, item.OrderAmount)
		totalHandlingFee = utils.FloatAdd(totalHandlingFee, item.TransferHandlingFee)
		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		w.SetRow(cell, row)
	}

	//totalAmount := xlsx.SetCellFormula(sheetName, "G"+strconv.Itoa(len(resp.List)+2), "=SUM(G2:"+"G"+strconv.Itoa(len(resp.List)+1)+")")
	//totalHandlingFee := xlsx.SetCellFormula(sheetName, "H"+strconv.Itoa(len(resp.List)+2), "=SUM(H2:"+"H"+strconv.Itoa(len(resp.List)+1)+")")
	row := []interface{}{
		excelize.Cell{StyleID: bottomStyle, Value: i18n.Sprintf("count")}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: totalAmount},
		excelize.Cell{StyleID: bottomStyle, Value: totalHandlingFee},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
		excelize.Cell{StyleID: bottomStyle, Value: ""}, excelize.Cell{StyleID: bottomStyle, Value: ""},
	}
	footerRow := strconv.Itoa(len(resp.List) + 2)
	w.SetRow("A"+footerRow, row)
	w.Flush()

	return
}
