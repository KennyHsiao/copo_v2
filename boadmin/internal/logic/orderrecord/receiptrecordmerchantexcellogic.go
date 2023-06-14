package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/excelizeutil"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/xuri/excelize/v2"
	"strconv"
	"sync"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReceiptRecordMerchantExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReceiptRecordMerchantExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) ReceiptRecordMerchantExcelLogic {
	return ReceiptRecordMerchantExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReceiptRecordMerchantExcelLogic) ReceiptRecordMerchantExcel(req *types.ReceiptRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	// 設置i18n
	utils.SetI18n(req.Language)

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}

	sheetName := i18n.Sprintf("Receipt Report")
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
		resp, err1 = orderrecordService.ReceiptRecordQueryAll(l.svcCtx.MyDB, *req, true, nil)
	}()
	go func() {
		defer wg.Done()
		totalResp, err2 = orderrecordService.ReceiptRecordTotalInfoBySuccess(l.svcCtx.MyDB, *req, nil)
	}()
	go func() {
		defer wg.Done()
		orderAmount, err3 = orderrecordService.ReceiptRecordTotalOrderAmount(l.svcCtx.MyDB, *req, nil)
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
	// 設置頁腳 Style
	footerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#ffff99"},
			Pattern: 1,
		},
	})

	// 設置標題
	header := []interface{}{}
	if req.ReportType == "1" {
		header = append(header, []interface{}{
			i18n.Sprintf("Merchant number"), i18n.Sprintf("Agent number"),
		}...)
	}
	header = append(header, []interface{}{
		i18n.Sprintf("Merchant order number"), i18n.Sprintf("Platform order number"),
		i18n.Sprintf("Receipt method"), i18n.Sprintf("Payment types"),
		i18n.Sprintf("Amount received"), i18n.Sprintf("Net amount received"),
		i18n.Sprintf("Actual handling fee"), i18n.Sprintf("Available amount"),
		i18n.Sprintf("Merchant rate"), i18n.Sprintf("Merchant handling fee"),
		i18n.Sprintf("Agent commission"),
		i18n.Sprintf("Order status"), i18n.Sprintf("Call back status"),
		i18n.Sprintf("Cause"), i18n.Sprintf("Note"),
		i18n.Sprintf("Bill of lading Time"), i18n.Sprintf("Transaction time"),
		i18n.Sprintf("Call back time"), i18n.Sprintf("Currency"),
	}...)
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {

		if item.Status != constants.SUCCESS {
			item.ActualAmount = 0
			item.TransferHandlingFee = 0
			item.TransferAmount = 0
		}
		row := []interface{}{}
		if req.ReportType == "1" {
			row = append(row, []interface{}{
				item.MerchantCode, item.AgentParentCode,
			}...)
		}
		row = append(row, []interface{}{
			item.MerchantOrderNo, item.OrderNo,
			excelizeutil.GetTxOrderTypeName(item.Type), item.PayTypeName,
			item.OrderAmount, item.ActualAmount,
			item.TransferHandlingFee, item.TransferAmount,
			item.Fee, item.HandlingFee,
			item.ProfitAmount,
			excelizeutil.GetTxOrderStatusName(item.Status), excelizeutil.GetTxMerchantCallbackName(item.IsMerchantCallback),
			excelizeutil.GetTxOrderReasonType(item.ReasonType), item.Memo,
			item.CreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.MerchantCallBackAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.CurrencyCode,
		}...)

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		// 把资料塞入 row
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 設置總金額
	totalRow := []interface{}{}
	if req.ReportType == "1" {
		totalRow = append(totalRow, []interface{}{
			i18n.Sprintf("count"), "", "",
		}...)
	} else {
		totalRow = append(totalRow, i18n.Sprintf("count"))
	}
	totalRow = append(totalRow, []interface{}{
		"",
		"", "",
		totalResp.TotalOrderAmount, totalResp.TotalActualAmount,
		totalResp.TotalTransferHandlingFee, totalResp.TotalTransferAmount,
		"", "",
		"", "",
		"", "",
		"", "",
		"", "",
	}...)
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &totalRow); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "T1", headerStyle)
	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "T"+footerRow, footerStyle)

	return
}
