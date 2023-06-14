package orderrecord

import (
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/common/excelizeutil"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/xuri/excelize/v2"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductRecordMerchantExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeductRecordMerchantExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) DeductRecordMerchantExcelLogic {
	return DeductRecordMerchantExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeductRecordMerchantExcelLogic) DeductRecordMerchantExcel(req types.DeductRecordQueryAllRequestX) (xlsx *excelize.File, err error) {
	// 設置i18n
	utils.SetI18n(req.Language)

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}

	sheetName := i18n.Sprintf("Payment Report")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.DeductRecordQueryAllResponseX
	if resp, err = orderrecordService.DeductRecordQueryAll(l.svcCtx.MyDB, req, true, l.ctx); err != nil {
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
	header := []interface{}{}
	if req.ReportType == "1" {
		header = append(header, []interface{}{
			i18n.Sprintf("Merchant number"), i18n.Sprintf("Agent number"),
		}...)
	}
	header = append(header, []interface{}{
		i18n.Sprintf("Merchant order number"), i18n.Sprintf("Platform order number"),
		i18n.Sprintf("Bank"),
		i18n.Sprintf("Branch province"), i18n.Sprintf("Branch city"),
		i18n.Sprintf("Account name"), i18n.Sprintf("Bank card number"),
		i18n.Sprintf("Payment method"), i18n.Sprintf("Payment amount"),
		i18n.Sprintf("Merchant rate"), i18n.Sprintf("Merchant handling fee"),
		i18n.Sprintf("Agent commission"),
		i18n.Sprintf("Order status"), i18n.Sprintf("Note"),
		i18n.Sprintf("Bill of lading Time"), i18n.Sprintf("Transaction time"),
		i18n.Sprintf("Order source"), i18n.Sprintf("Currency"),
	}...)

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
		row := []interface{}{}
		if req.ReportType == "1" {
			row = append(row, []interface{}{
				item.MerchantCode, item.AgentParentCode,
			}...)
		}
		row = append(row, []interface{}{
			item.MerchantOrderNo, orderNo,
			item.MerchantBankName,
			item.MerchantBankProvince, item.MerchantBankCity,
			item.MerchantAccountName, item.MerchantBankAccount,
			excelizeutil.GetTxOrderTypeName(item.Type), item.OrderAmount,
			item.Fee, item.HandlingFee,
			item.ProfitAmount,
			excelizeutil.GetTxOrderStatusName(item.Status), memo,
			createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			excelizeutil.GetTxOrderSourceName(item.Source), item.CurrencyCode,
		}...)

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	row := []interface{}{i18n.Sprintf("count")}
	if req.ReportType == "1" {
		xlsx.SetCellFormula(sheetName, "K"+strconv.Itoa(len(resp.List)+2), "=SUM(K2:"+"K"+strconv.Itoa(len(resp.List)+1)+")")
		xlsx.SetCellFormula(sheetName, "M"+strconv.Itoa(len(resp.List)+2), "=SUM(M2:"+"M"+strconv.Itoa(len(resp.List)+1)+")")
	} else {
		xlsx.SetCellFormula(sheetName, "I"+strconv.Itoa(len(resp.List)+2), "=SUM(I2:"+"I"+strconv.Itoa(len(resp.List)+1)+")")
		xlsx.SetCellFormula(sheetName, "K"+strconv.Itoa(len(resp.List)+2), "=SUM(K2:"+"K"+strconv.Itoa(len(resp.List)+1)+")")
	}

	// 設置row 高度
	xlsx.SetRowHeight(sheetName, len(resp.List)+2, rowHeight)
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &row); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)

	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "Q1", headerStyle)

	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "Q"+footerRow, bottomStyle)

	return
}
