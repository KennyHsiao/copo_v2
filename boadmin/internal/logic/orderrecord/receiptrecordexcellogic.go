package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/downloadReportService"
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/excelizeutil"
	"com.copo/bo_service/common/utils"
	"context"
	"encoding/json"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
	"sync"
)

type ReceiptRecordExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReceiptRecordExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReceiptRecordExcelLogic {
	return ReceiptRecordExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReceiptRecordExcelLogic) ReceiptRecordExcel(req *types.ReceiptRecordQueryAllRequestX) (err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	userId := l.ctx.Value("userId").(json.Number)
	userIdint, _ := userId.Int64()
	isAdmin := l.ctx.Value("isAdmin").(bool)
	requestBytes, _ := json.Marshal(req)

	createDownloadTask := &types.CreateDownloadTask{
		Prefix:       "收款纪录数据",
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
	downReportID, fileName, err := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
	if err != nil {
		return err
	}
	go func() {
		xlsx, err := newGenerateReceiptExcel(l.svcCtx.MyDB, req)
		if err != nil {
			logx.WithContext(l.ctx).Error("newGenerateReceiptExcel Err: " + err.Error())
		}
		downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
	}()

	return
}

func generateReceiptExcelOld(db *gorm.DB, req *types.ReceiptRecordQueryAllRequestX) (xlsx *excelize.File, err error) {

	// 設置i18n
	utils.SetI18n(req.Language)

	sheetName := "收款记录(后台)"
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.ReceiptRecordQueryAllResponseX
	var totalResp *types.ReceiptRecordTotalInfoResponse
	var orderAmount float64
	var err1 error
	var err2 error
	var err3 error

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		resp, err1 = orderrecordService.ReceiptRecordQueryAll(db, *req, true, nil)
	}()
	go func() {
		defer wg.Done()
		totalResp, err2 = orderrecordService.ReceiptRecordTotalInfoBySuccess(db, *req, nil)
	}()
	go func() {
		defer wg.Done()
		orderAmount, err3 = orderrecordService.ReceiptRecordTotalOrderAmount(db, *req, nil)
	}()

	wg.Wait()

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if err3 != nil {
		return nil, err3
	}
	totalResp.TotalOrderAmount = orderAmount

	// 建立excel
	xlsx = excelize.NewFile()
	// 將預設分頁更名
	xlsx.SetSheetName("Sheet1", sheetName)

	// 設置標頭 Style
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	// 設置資料 Style
	footerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 設置標題
	header := []interface{}{
		"平台订单号", "渠道订单号",
		"商户订单号", "渠道名称",
		"渠道编号", "商户编号",
		"银行后五码", "收款方式",
		"支付类型", "收款户名",
		"收款账号", "打款户名",
		"打款账号", "收款金额",
		"实付金额", "实收手续费",
		"可用金额", "商户费率",
		"商户手续费", "订单状态",
		"备注", "原因",
		"提单时间", "交易时间",
		"回调时间", "币别",
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {
		var orderNo = item.OrderNo
		if item.IsTest == "1" {
			orderNo = item.OrderNo + "(测)"
		}
		if item.Status != constants.SUCCESS {
			item.ActualAmount = 0
			item.TransferHandlingFee = 0
			item.TransferAmount = 0
		}
		row := []interface{}{
			orderNo, item.ChannelOrderNo,
			item.MerchantOrderNo, item.ChannelName,
			item.ChannelCode, item.MerchantCode,
			item.MerchantBankAccountLastFive, excelizeutil.GetTxOrderTypeName(item.Type),
			item.PayTypeName, item.ChannelAccountName,
			item.ChannelBankAccount, item.MerchantAccountName,
			item.MerchantBankAccount, item.OrderAmount,
			item.ActualAmount, item.TransferHandlingFee,
			item.TransferAmount, item.Fee,
			item.HandlingFee, excelizeutil.GetTxOrderStatusName(item.Status),
			item.Memo, excelizeutil.GetTxOrderReasonType(item.ReasonType),
			item.CreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.MerchantCallBackAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.CurrencyCode,
		}

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		// 塞資料
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 設置總金額
	totalRow := []interface{}{
		"加总", "",
		"", "",
		"", "",
		"", "",
		"", "",
		"", "",
		"", totalResp.TotalOrderAmount,
		totalResp.TotalActualAmount, totalResp.TotalTransferAmount,
		totalResp.TotalTransferHandlingFee, "",
		"", "",
		"", "",
		"", "",
		"", "",
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &totalRow); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "Z1", headerStyle)
	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "Z"+footerRow, footerStyle)

	return
}

func newGenerateReceiptExcel(db *gorm.DB, req *types.ReceiptRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	// 設置i18n
	utils.SetI18n(req.Language)

	sheetName := "收款记录(后台)"
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.ReceiptRecordQueryAllResponseX
	var totalResp *types.ReceiptRecordTotalInfoResponse
	var orderAmount float64
	var err1 error
	var err2 error
	var err3 error

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		resp, err1 = orderrecordService.ReceiptRecordQueryAll(db, *req, true, nil)
	}()
	go func() {
		defer wg.Done()
		totalResp, err2 = orderrecordService.ReceiptRecordTotalInfoBySuccess(db, *req, nil)
	}()
	go func() {
		defer wg.Done()
		orderAmount, err3 = orderrecordService.ReceiptRecordTotalOrderAmount(db, *req, nil)
	}()

	wg.Wait()

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}
	if err3 != nil {
		return nil, err3
	}
	totalResp.TotalOrderAmount = orderAmount

	// 建立excel
	xlsx = excelize.NewFile()
	xlsx.SetSheetName("Sheet1", sheetName)
	w, err := xlsx.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	// 設置資料 Style
	footerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 設置標題
	header := []interface{}{
		"平台订单号", "渠道订单号",
		"商户订单号", "渠道名称",
		"渠道编号", "商户编号",
		"银行后五码", "收款方式",
		"支付类型", "收款户名",
		"收款账号", "提单姓名",
		"打款账号", "收款金额",
		"实付金额", "实收手续费",
		"可用金额", "商户费率",
		"商户手续费", "订单状态",
		"备注", "原因",
		"提单时间", "交易时间",
		"回调时间", "币别",
	}

	w.SetRow("A1", header,
		excelize.RowOpts{StyleID: headerStyle, Height: rowHeight, Hidden: false})

	// 迴圈建置資料
	for i, item := range resp.List {
		var orderNo = item.OrderNo
		if item.IsTest == "1" {
			orderNo = item.OrderNo + "(测)"
		}
		if item.Status != constants.SUCCESS {
			item.ActualAmount = 0
			item.TransferHandlingFee = 0
			item.TransferAmount = 0
		}
		row := []interface{}{
			orderNo, item.ChannelOrderNo,
			item.MerchantOrderNo, item.ChannelName,
			item.ChannelCode, item.MerchantCode,
			item.MerchantBankAccountLastFive, excelizeutil.GetTxOrderTypeName(item.Type),
			item.PayTypeName, item.ChannelAccountName,
			item.ChannelBankAccount, item.MerchantAccountName,
			item.MerchantBankAccount, item.OrderAmount,
			item.ActualAmount, item.TransferHandlingFee,
			item.TransferAmount, item.Fee,
			item.HandlingFee, excelizeutil.GetTxOrderStatusName(item.Status),
			item.Memo, excelizeutil.GetTxOrderReasonType(item.ReasonType),
			item.CreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.MerchantCallBackAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.CurrencyCode,
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		w.SetRow(cell, row)
	}
	// 設置總金額
	totalRow := []interface{}{
		excelize.Cell{StyleID: footerStyle, Value: "加总"}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: totalResp.TotalOrderAmount},
		excelize.Cell{StyleID: footerStyle, Value: totalResp.TotalActualAmount}, excelize.Cell{StyleID: footerStyle, Value: totalResp.TotalTransferAmount},
		excelize.Cell{StyleID: footerStyle, Value: totalResp.TotalTransferHandlingFee}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
		excelize.Cell{StyleID: footerStyle, Value: ""}, excelize.Cell{StyleID: footerStyle, Value: ""},
	}
	footerRow := strconv.Itoa(len(resp.List) + 2)
	w.SetRow("A"+footerRow, totalRow)
	w.Flush()

	return
}
