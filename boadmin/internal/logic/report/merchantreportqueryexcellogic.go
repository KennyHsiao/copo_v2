package report

import (
	"com.copo/bo_service/boadmin/internal/service/downloadReportService"
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/excelizeutil"
	"context"
	"encoding/json"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
)

type MerchantReportQueryExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantReportQueryExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantReportQueryExcelLogic {
	return MerchantReportQueryExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantReportQueryExcelLogic) MerchantReportQueryExcel(req *types.MerchantReportQueryRequestX) (err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	userId := l.ctx.Value("userId").(json.Number)
	userIdint, _ := userId.Int64()
	isAdmin := l.ctx.Value("isAdmin").(bool)
	requestBytes, _ := json.Marshal(req)

	createDownloadTask := &types.CreateDownloadTask{
		Prefix:       "商户报表数据",
		Infix:        "",
		Suffix:       "",
		IsAdmin:      isAdmin,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		CurrencyCode: req.CurrencyCode,
		MerchantCode: merchantCode,
		UserId:       userIdint,
		ReqParam:     string(requestBytes),
		Type:         constants.MERCHANT_REPORT,
	}
	if len(req.MerchantCode) > 0 && len(req.ChannelName) > 0 {
		createDownloadTask.Infix = req.MerchantCode + " | " + req.ChannelName + " | "
	} else if len(req.MerchantCode) > 0 && len(req.ChannelName) == 0 {
		createDownloadTask.Infix = req.MerchantCode + " | "
	} else if len(req.MerchantCode) == 0 && len(req.ChannelName) > 0 {
		createDownloadTask.Infix = req.ChannelName + " | "
	}
	downReportID, fileName, err := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
	if err != nil {
		return err
	}
	go func() {
		xlsx, err := newGenerateMerchantExcel(l.svcCtx.MyDB, req)
		if err != nil {
			logx.WithContext(l.ctx).Error("newGenerateMerchantExcel Err: " + err.Error())
		}
		downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
	}()

	return
}
func generateMerchantExcel(db *gorm.DB, req *types.MerchantReportQueryRequestX) (xlsx *excelize.File, err error) {
	sheetName := "商戶报表"
	var rowHeight float64 = 17

	//取得資料
	var reportResp *types.MerchantReportQueryResponse
	req.PageSize = 0
	if reportResp, err = reportService.InterMerchantReport(db, req, nil); err != nil {
		return
	}

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
		"商户編号", "渠道編码",
		"渠道名称", "交易类型",
		"支付类型", "订单总数",
		"订单总额", "成功总数",
		"成功总额", "成功率",
		"渠道费率", "商户费率",
		"渠道手续费", "商户手续费",
		"系统成本", "系统盈利",
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range reportResp.List {
		row := []interface{}{
			item.MerchantCode, item.ChannelCode,
			item.ChannelName, item.TransactionType,
			item.PayTypeName, item.OrderQuantity,
			item.OrderAmount, item.SuccessQuantity,
			item.SuccessAmount, item.SuccessRate,
			item.ChannelFee, item.MerchantFee,
			item.ChannelHandlingFee, item.MerchantHandlingFee,
			item.SystemCost, item.SystemProfit,
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
		"加总", "--",
		"--", "--",
		"--", reportResp.TotalOrderQuantity,
		reportResp.TotalOrderAmount, reportResp.TotalSuccessQuantity,
		reportResp.TotalSuccessAmount, "--",
		"--", "--",
		"--", "--",
		reportResp.TotalCost, reportResp.TotalProfit,
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(reportResp.List)+2), &totalRow); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "G1", headerStyle)
	// 設置頁腳 style
	footerRow := strconv.Itoa(len(reportResp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "P"+footerRow, footerStyle)

	return
}

func newGenerateMerchantExcel(db *gorm.DB, req *types.MerchantReportQueryRequestX) (xlsx *excelize.File, err error) {
	sheetName := "商戶报表"
	var rowHeight float64 = 17

	//取得資料
	var reportResp *types.MerchantReportQueryResponse
	req.PageSize = 0
	if reportResp, err = reportService.InterMerchantReport(db, req, nil); err != nil {
		return
	}

	// 建立excel
	xlsx = excelize.NewFile()
	// 將預設分頁更名
	xlsx.SetSheetName("Sheet1", sheetName)
	w, err := xlsx.NewStreamWriter(sheetName)
	if err != nil {
		return nil, err
	}
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
		"商户編号", "渠道編码",
		"渠道名称", "交易类型",
		"支付类型", "订单总数",
		"订单总额", "成功总数",
		"成功总额", "成功率",
		"渠道费率", "商户费率",
		"渠道手续费", "商户手续费",
		"系统成本", "系统盈利",
	}
	w.SetRow("A1", header,
		excelize.RowOpts{StyleID: headerStyle, Height: rowHeight, Hidden: false})

	// 迴圈建置資料
	for i, item := range reportResp.List {
		row := []interface{}{
			item.MerchantCode, item.ChannelCode,
			item.ChannelName, item.TransactionType,
			item.PayTypeName, item.OrderQuantity,
			item.OrderAmount, item.SuccessQuantity,
			item.SuccessAmount, item.SuccessRate,
			item.ChannelFee, item.MerchantFee,
			item.ChannelHandlingFee, item.MerchantHandlingFee,
			item.SystemCost, item.SystemProfit,
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		w.SetRow(cell, row)
	}

	// 設置總金額
	totalRow := []interface{}{
		excelize.Cell{StyleID: footerStyle, Value: "加总"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalOrderQuantity},
		excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalOrderAmount}, excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalOrderAmount},
		excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalSuccessAmount}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalCost}, excelize.Cell{StyleID: footerStyle, Value: reportResp.TotalProfit},
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(reportResp.List)+2), &totalRow); err != nil {
		return
	}

	footerRow := strconv.Itoa(len(reportResp.List) + 2)
	w.SetRow("A"+footerRow, totalRow)
	w.Flush()

	return
}
