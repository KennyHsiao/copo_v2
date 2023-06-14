package commissionMonthReport

import (
	"archive/zip"
	"bytes"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	_ "image/png"
	"strconv"
)

type MonthReportExcelLogic struct {
	logx.Logger
	ctx        context.Context
	svcCtx     *svc.ServiceContext
	payTypeMap map[string]string
}

func NewMonthReportExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportExcelLogic {
	return MonthReportExcelLogic{
		Logger:     logx.WithContext(ctx),
		ctx:        ctx,
		svcCtx:     svcCtx,
		payTypeMap: make(map[string]string),
	}
}

func (l *MonthReportExcelLogic) MonthReportExcel(req *types.CommissionMonthReportExcelRequest) (*bytes.Buffer, error) {

	var xlsxList []*excelize.File

	// 取得支付類型對應Map
	if err := l.getPayTypeMap(); err != nil {
		return nil, err
	}

	// 迴圈產生多個EXCEL
	for _, id := range req.IDList {
		xlsx, err := l.CreateExcel(id)
		if err != nil {
			return nil, err
		}
		xlsxList = append(xlsxList, xlsx)
	}

	// 把EXCEL加入壓縮檔
	return l.compressXlsx(xlsxList)
}

func (l *MonthReportExcelLogic) CreateExcel(id int64) (xlsx *excelize.File, err error) {
	var report types.CommissionMonthReportX
	var subreports []types.CommissionMonthReportX

	// 取得報表
	if err = l.svcCtx.MyDB.Table("cm_commission_month_reports").Take(&report, id).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = l.svcCtx.MyDB.Table("cm_commission_month_reports").
		Where("agent_layer_no != ?", report.AgentLayerNo).
		Where("agent_layer_no like ?", report.AgentLayerNo+"%").
		Where("month = ?", report.Month).
		Where("currency_code = ?", report.CurrencyCode).
		Order("agent_layer_no asc").
		Find(&subreports).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	sheetName := report.MerchantCode
	if report.CurrencyCode != "CNY" {
		sheetName += "(" + report.CurrencyCode + ")"
	}
	// 建立excel
	xlsx = excelize.NewFile()
	// 將預設分頁更名
	xlsx.SetSheetName("Sheet1", sheetName)

	// 自身報表
	l.insertAgentData(report, xlsx, sheetName, 1, 1)
	// 迴圈跑子報表
	for i, subreport := range subreports {
		//层级编号每差三位字 = 多一层级
		level := (len(subreport.AgentLayerNo)-len(report.AgentLayerNo))/3 + 1
		// 資料開始欄 = 1(初始位置) + 第几个子报表 * 8
		startColumn := 1 + (i+1)*8
		l.insertAgentData(subreport, xlsx, sheetName, startColumn, level)
	}
	if report.ChangeCommission > 0 {
		l.CreateChangeRecordSheet(xlsx, report.MerchantCode)
	}

	return
}

func (l *MonthReportExcelLogic) insertAgentData(report types.CommissionMonthReportX, xlsx *excelize.File, sheetName string, startCol int, level int) (err error) {

	var merchantCodeList []string

	chineseNumbers := []string{
		"零", "一", "二", "三", "四", "五", "六", "七", "八", "九", "十",
	}

	startColumn, _ := excelize.ColumnNumberToName(startCol)
	endColumn, _ := excelize.ColumnNumberToName(startCol + 7)
	dataStartColumn, _ := excelize.ColumnNumberToName(startCol + 1)
	dataEndColumn, _ := excelize.ColumnNumberToName(startCol + 6)

	// 設置欄位寬度
	xlsx.SetColWidth(sheetName, endColumn, dataEndColumn, 8)
	xlsx.SetColWidth(sheetName, dataStartColumn, dataEndColumn, 15)
	xlsx.SetColWidth(sheetName, dataStartColumn, dataStartColumn, 4.5)

	// 插入圖片
	if err = xlsx.AddPicture(sheetName, dataStartColumn+"2", "etc/resources/images/copoLog.png", `{
        "x_scale": 0.5,
        "y_scale": 0.5
    }`); err != nil {
		logx.Error(err)
	}

	// 取得所有拥金来源商户号 Distinct
	if merchantCodeList, err = l.getMerchantCodeList(report.ID); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 设置开始月份
	monthStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
	})
	xlsx.SetCellValue(sheetName, dataEndColumn+"2", fmt.Sprintf("计算月份：%s", report.Month))
	xlsx.SetCellStyle(sheetName, dataEndColumn+"2", dataEndColumn+"2", monthStyle)
	xlsx.SetCellValue(sheetName, dataEndColumn+"3", fmt.Sprintf("币别：%s", report.CurrencyCode))
	xlsx.SetCellStyle(sheetName, dataEndColumn+"3", dataEndColumn+"2", monthStyle)

	// 设置代理商户号标题
	agentMerchantCodeStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Color: "#FF8000",
			Bold:  true,
			Size:  18,
		},
	})
	xlsx.MergeCell(sheetName, startColumn+"4", endColumn+"4")
	xlsx.SetCellValue(sheetName, startColumn+"4", report.MerchantCode)
	xlsx.SetCellStyle(sheetName, startColumn+"4", endColumn+"4", agentMerchantCodeStyle)
	xlsx.SetRowHeight(sheetName, 4, 24)

	// 设置层级文字
	levelStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
		Font: &excelize.Font{
			Bold: true,
		},
	})
	xlsx.SetCellValue(sheetName, dataStartColumn+"5", fmt.Sprintf("第%s级代理", chineseNumbers[level]))
	xlsx.SetCellStyle(sheetName, dataStartColumn+"5", dataStartColumn+"5", levelStyle)

	// 设置佣金资料
	dataRow := 6 //资料开始行
	for _, merchantCode := range merchantCodeList {
		// 取得单商户佣金资料 並回傳用到第幾行
		if dataRow, err = l.setMerchantData(xlsx, sheetName, report.ID, merchantCode, dataRow, startCol); err != nil {
			return
		}
	}

	//设置总计页脚
	footColumnA1, _ := excelize.ColumnNumberToName(startCol + 1)
	footColumnA2, _ := excelize.ColumnNumberToName(startCol + 4)
	footColumnB, _ := excelize.ColumnNumberToName(startCol + 5)
	footColumnC, _ := excelize.ColumnNumberToName(startCol + 6)

	topBorder := excelize.Border{Type: "top", Style: 1, Color: "333333"}
	leftBorder := excelize.Border{Type: "left", Style: 1, Color: "333333"}
	rightBorder := excelize.Border{Type: "right", Style: 1, Color: "333333"}
	bottomBorder := excelize.Border{Type: "bottom", Style: 1, Color: "333333"}

	footHeaderStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})
	footValueStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})
	totalAmount := report.PayTotalAmount + report.InternalChargeTotalAmount + report.ProxyPayTotalAmount
	// 設置總計
	if report.ChangeCommission > 0 {

		xlsx.MergeCell(sheetName, footColumnA1+strconv.Itoa(dataRow), footColumnA2+strconv.Itoa(dataRow))
		xlsx.SetCellValue(sheetName, footColumnA1+strconv.Itoa(dataRow), "小计")
		xlsx.SetCellValue(sheetName, footColumnB+strconv.Itoa(dataRow), totalAmount)
		xlsx.SetCellValue(sheetName, footColumnC+strconv.Itoa(dataRow), report.TotalCommission)

		xlsx.MergeCell(sheetName, footColumnA1+strconv.Itoa(dataRow+1), footColumnA2+strconv.Itoa(dataRow+1))
		xlsx.SetCellValue(sheetName, footColumnA1+strconv.Itoa(dataRow+1), "调整异动")
		xlsx.SetCellValue(sheetName, footColumnB+strconv.Itoa(dataRow+1), "--")
		xlsx.SetCellValue(sheetName, footColumnC+strconv.Itoa(dataRow+1), report.ChangeCommission-report.TotalCommission)

		xlsx.MergeCell(sheetName, footColumnA1+strconv.Itoa(dataRow+2), footColumnA2+strconv.Itoa(dataRow+2))
		xlsx.SetCellValue(sheetName, footColumnA1+strconv.Itoa(dataRow+2), "总计")
		xlsx.SetCellValue(sheetName, footColumnB+strconv.Itoa(dataRow+2), totalAmount)
		xlsx.SetCellValue(sheetName, footColumnC+strconv.Itoa(dataRow+2), report.ChangeCommission)

		// 設置Style
		xlsx.SetCellStyle(sheetName, footColumnA1+strconv.Itoa(dataRow), footColumnA2+strconv.Itoa(dataRow+2), footHeaderStyle)
		xlsx.SetCellStyle(sheetName, footColumnB+strconv.Itoa(dataRow), footColumnC+strconv.Itoa(dataRow+2), footValueStyle)
	} else {
		xlsx.MergeCell(sheetName, footColumnA1+strconv.Itoa(dataRow), footColumnA2+strconv.Itoa(dataRow))
		xlsx.SetCellValue(sheetName, footColumnA1+strconv.Itoa(dataRow), "总计")
		xlsx.SetCellValue(sheetName, footColumnB+strconv.Itoa(dataRow), totalAmount)
		xlsx.SetCellValue(sheetName, footColumnC+strconv.Itoa(dataRow), report.TotalCommission)

		// 設置Style
		xlsx.SetCellStyle(sheetName, footColumnA1+strconv.Itoa(dataRow), footColumnA2+strconv.Itoa(dataRow), footHeaderStyle)
		xlsx.SetCellStyle(sheetName, footColumnB+strconv.Itoa(dataRow), footColumnC+strconv.Itoa(dataRow), footValueStyle)
	}

	return
}

func (l *MonthReportExcelLogic) getMerchantCodeList(reportId int64) (merchantCodeList []string, err error) {
	err = l.svcCtx.MyDB.Table("cm_commission_month_report_details").
		Select("merchant_code").
		Where("commission_month_report_id = ? ", reportId).
		Order("merchant_code asc").
		Distinct().Pluck("merchant_code", &merchantCodeList).Error
	return
}

// 取得单商户佣金资料 並回傳用到第幾行
func (l *MonthReportExcelLogic) setMerchantData(xlsx *excelize.File, sheetName string, reportId int64, merchantCode string, row int, startCol int) (endRow int, err error) {
	var feeDetails []types.CommissionMonthReportDetailX
	var handlingFeeDetails []types.CommissionMonthReportDetailX

	topBorder := excelize.Border{Type: "top", Style: 1, Color: "333333"}
	leftBorder := excelize.Border{Type: "left", Style: 1, Color: "333333"}
	rightBorder := excelize.Border{Type: "right", Style: 1, Color: "333333"}
	bottomBorder := excelize.Border{Type: "bottom", Style: 1, Color: "333333"}

	// 取得 支付&内充（收费率的代付） 佣金资料
	if err = l.svcCtx.MyDB.Table("cm_commission_month_report_details").
		Where("commission_month_report_id = ? ", reportId).
		Where("merchant_code = ? ", merchantCode).
		Where("(order_type in ('ZF','NC') OR (order_type = 'DF' AND diff_fee > 0) ) ").
		Order("order_type desc").
		Order("pay_type_code asc").
		Order("merchant_fee asc").
		Find(&feeDetails).Error; err != nil {
		return endRow, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 取得 代付(只收手续费) 佣金资料
	if err = l.svcCtx.MyDB.Table("cm_commission_month_report_details").
		Where("commission_month_report_id = ? ", reportId).
		Where("merchant_code = ? ", merchantCode).
		Where("order_type = 'DF' AND diff_fee = 0 ").
		Order("diff_handling_fee asc").
		Find(&handlingFeeDetails).Error; err != nil {
		return endRow, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	dataStartColumn, _ := excelize.ColumnNumberToName(startCol + 1)
	payTypeColumn, _ := excelize.ColumnNumberToName(startCol + 2)
	dataEndColumn, _ := excelize.ColumnNumberToName(startCol + 6)

	// 设置下层商户号标题
	merchantCodeStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
		Font: &excelize.Font{
			Bold: false,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#F0F0F0"},
			Pattern: 1,
		},
	})
	xlsx.MergeCell(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row))
	xlsx.SetCellValue(sheetName, dataStartColumn+strconv.Itoa(row), fmt.Sprintf("商戶编号：%s", merchantCode))
	xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row), merchantCodeStyle)
	row += 1

	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Font: &excelize.Font{
			Bold: true,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})
	valueCenterStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})
	valueStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "right",
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})

	if len(feeDetails) > 0 {
		// 設置標題
		zfHeader := []interface{}{
			"序", "支付类型", "商户费率", "代理费率", "交易金额", "代理佣金",
		}
		if err = xlsx.SetSheetRow(sheetName, dataStartColumn+strconv.Itoa(row), &zfHeader); err != nil {
			return
		}
		xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row), headerStyle)
		row += 1

		for index, item := range feeDetails {
			rowData := []interface{}{
				strconv.Itoa(index + 1), l.payTypeMap[item.PayTypeCode],
				fmt.Sprintf("%.2f", item.MerchantFee), fmt.Sprintf("%.2f", item.AgentFee),
				item.TotalAmount, item.TotalCommission,
			}
			// 把资料塞入
			if err = xlsx.SetSheetRow(sheetName, dataStartColumn+strconv.Itoa(row), &rowData); err != nil {
				return
			}
			xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row), valueStyle)
			xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), payTypeColumn+strconv.Itoa(row), valueCenterStyle)
			row += 1
		}
		row += 1 // 空一行
	}

	if len(handlingFeeDetails) > 0 {
		// 設置標題
		zfHeader := []interface{}{
			"序", "交易类型", "笔数", "单笔手续费", "交易金额", "代理佣金",
		}
		if err = xlsx.SetSheetRow(sheetName, dataStartColumn+strconv.Itoa(row), &zfHeader); err != nil {
			return
		}
		xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row), headerStyle)
		row += 1
		for index, item := range handlingFeeDetails {
			rowData := []interface{}{
				strconv.Itoa(index + 1), l.payTypeMap[item.PayTypeCode],
				fmt.Sprintf("%.2f", item.TotalNumber), fmt.Sprintf("%.2f", item.DiffHandlingFee),
				item.TotalAmount, item.TotalCommission,
			}
			// 把资料塞入
			if err = xlsx.SetSheetRow(sheetName, dataStartColumn+strconv.Itoa(row), &rowData); err != nil {
				return
			}
			xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), dataEndColumn+strconv.Itoa(row), valueStyle)
			xlsx.SetCellStyle(sheetName, dataStartColumn+strconv.Itoa(row), payTypeColumn+strconv.Itoa(row), valueCenterStyle)
			row += 1
		}
		row += 1 // 空一行
	}

	return row, err
}

func (l *MonthReportExcelLogic) getPayTypeMap() (err error) {
	var payTypes []types.PayType
	payTypeMap := make(map[string]string)
	if err = l.svcCtx.MyDB.Table("ch_pay_types").Find(&payTypes).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	payTypeMap["NC"] = "內充"
	for _, payType := range payTypes {
		payTypeMap[payType.Code] = payType.Name
	}
	l.payTypeMap = payTypeMap
	return
}

func (l *MonthReportExcelLogic) compressXlsx(xlsxList []*excelize.File) (*bytes.Buffer, error) {
	// 建立一個緩衝區用來儲存壓縮檔案內容
	buf := new(bytes.Buffer)
	// 建立一個壓縮檔案
	zipWriter := zip.NewWriter(buf)

	for _, xlsx := range xlsxList {
		f, err := zipWriter.Create(xlsx.GetSheetName(0) + ".xlsx")
		if err != nil {
			return nil, err
		}
		if err = xlsx.Write(f); err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		logx.Error(err)
	}
	return buf, nil
}

func (l *MonthReportExcelLogic) CreateChangeRecordSheet(xlsx *excelize.File, merchantCode string) (err error) {
	sheetName := "调整异动未计入系统流水"
	xlsx.NewSheet(sheetName)

	topBorder := excelize.Border{Type: "top", Style: 1, Color: "333333"}
	leftBorder := excelize.Border{Type: "left", Style: 1, Color: "333333"}
	rightBorder := excelize.Border{Type: "right", Style: 1, Color: "333333"}
	bottomBorder := excelize.Border{Type: "bottom", Style: 1, Color: "333333"}

	// 设置下层商户号标题
	merchantCodeStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: false,
			Size: 14,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#CFE2F3"},
			Pattern: 1,
		},
	})
	totalStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: false,
			Size: 11,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFF2CC"},
			Pattern: 1,
		},
	})
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})
	valueStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})

	// 大標題列
	xlsx.MergeCell(sheetName, "A1", "I1")
	xlsx.SetCellValue(sheetName, "A1", fmt.Sprintf("商戶号：%s 调整异动未计入系统流水", merchantCode))
	xlsx.SetCellStyle(sheetName, "A1", "I1", merchantCodeStyle)

	// 總計列
	xlsx.SetCellStyle(sheetName, "G2", "I2", totalStyle)
	xlsx.SetCellValue(sheetName, "G2", "小计")
	xlsx.SetCellFormula(sheetName, "H2", "=SUM(H4:H200)")
	xlsx.SetCellFormula(sheetName, "I2", "=SUM(I4:I200)")

	// 表格標題列
	headers := []interface{}{
		"序", "商户号", "单号", "支付类型", "备注", "商户价格", "代理价格", "订单金额", "代理佣金",
	}
	xlsx.SetCellStyle(sheetName, "A3", "I3", headerStyle)
	if err = xlsx.SetSheetRow(sheetName, "A3", &headers); err != nil {
		return
	}

	for i := 1; i <= 100; i++ {
		row := i + 3
		xlsx.SetCellValue(sheetName, "A"+strconv.Itoa(row), strconv.Itoa(i))
		xlsx.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), headerStyle)
		xlsx.SetCellStyle(sheetName, "B"+strconv.Itoa(row), "I"+strconv.Itoa(row), valueStyle)
		xlsx.SetCellFormula(sheetName, "I"+strconv.Itoa(i+3), fmt.Sprintf("=H%d*(F%d%%-G%d%%)", row, row, row))
	}

	xlsx.SetRowHeight(sheetName, 1, 30)
	xlsx.SetRowHeight(sheetName, 3, 30)
	xlsx.SetColWidth(sheetName, "B", "I", 15)
	xlsx.SetColWidth(sheetName, "A", "A", 4.5)
	xlsx.SetColWidth(sheetName, "C", "C", 30)

	return err
}
