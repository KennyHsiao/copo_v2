package report

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncomReportExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomReportExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomReportExcelLogic {
	return IncomReportExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomReportExcelLogic) IncomReportExcel(req *types.IncomReportMonthQueryRequest) (xlsx *excelize.File, err error) {
	var incomReport []types.IncomReport

	if err = l.svcCtx.MyDB.Table("rp_incom_report").
		Where("month >= ? AND month <= ?", req.StartMonth, req.EndMonth).
		Where("currency_code = ?", req.CurrencyCode).
		Find(&incomReport).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE)
	}

	sheetName := "收益报表"
	var rowHeight float64 = 17
	// 建立excel
	xlsx = excelize.NewFile()
	// 將預設分頁更名
	xlsx.SetSheetName("Sheet1", sheetName)

	dataEndColumn, _ := excelize.ColumnNumberToName(3 + len(incomReport) - 1)
	// 設置寬度
	xlsx.SetColWidth(sheetName, "C", dataEndColumn, 20)
	// 設置style
	topBorder := excelize.Border{Type: "top", Style: 1, Color: "333333"}
	leftBorder := excelize.Border{Type: "left", Style: 1, Color: "333333"}
	rightBorder := excelize.Border{Type: "right", Style: 1, Color: "333333"}
	bottomBorder := excelize.Border{Type: "bottom", Style: 1, Color: "333333"}

	style, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})

	numStyle, _ := xlsx.NewStyle(&excelize.Style{
		NumFmt: 39,
		Border: []excelize.Border{
			leftBorder, topBorder, rightBorder, bottomBorder,
		},
	})

	xlsx.SetCellStyle(sheetName, "C"+strconv.Itoa(4), dataEndColumn+strconv.Itoa(17), numStyle)
	xlsx.SetCellStyle(sheetName, "C"+strconv.Itoa(3), dataEndColumn+strconv.Itoa(3), style)

	xlsx.SetCellStyle(sheetName, "A3", "A17", style)
	xlsx.SetCellStyle(sheetName, "B3", "B17", style)

	// 標題
	xlsx, err = l.TitleInfo(xlsx, sheetName, incomReport, rowHeight)

	// 月份
	xlsx, err = l.MonthInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	// 支付
	xlsx, err = l.ZfInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}
	// 內充
	xlsx, err = l.NcInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	// 收款總收益
	xlsx, err = l.ReceivedTotalNetProfitInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	// 下發
	xlsx, err = l.WfInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}
	//代付
	xlsx, err = l.DfInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	//拨款总手续费
	xlsx, err = l.AllocInfo(xlsx, sheetName, incomReport, rowHeight)

	// 出款總收益
	xlsx, err = l.RemitTotalNetProfitInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	// 淨盈利總計
	xlsx, err = l.TotalNetProfitInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	// 代理佣金總計
	xlsx, err = l.CommissionTotalAmountInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}
	// 盈利成長率(%)
	xlsx, err = l.ProfitGrownRateInfo(xlsx, sheetName, incomReport, rowHeight)
	if err != nil {
		return
	}

	return xlsx, nil
}

func (l *IncomReportExcelLogic) ZfInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A4合并A5
	headerA4 := []interface{}{
		"支付",
	}
	if err := xlsx.SetSheetRow(sheetName, "A4", &headerA4); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A4", "A5"); err != nil {
		return nil, err
	}
	// B4
	headerB4 := []interface{}{
		"總流水",
	}
	if err := xlsx.SetSheetRow(sheetName, "B4", &headerB4); err != nil {
		return nil, err
	}
	// B5
	headerB5 := []interface{}{
		"淨盈利(A)",
	}
	if err := xlsx.SetSheetRow(sheetName, "B5", &headerB5); err != nil {
		return nil, err
	}
	row := []interface{}{}
	row2 := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.PayTotalAmount)
		row2 = append(row2, report.PayNetProfit)
	}

	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 4, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(4), &row); err != nil {
		return nil, err
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 5, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(5), &row2); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) NcInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A6合并A7
	headerA6 := []interface{}{
		"内充",
	}
	if err := xlsx.SetSheetRow(sheetName, "A6", &headerA6); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A6", "A7"); err != nil {
		return nil, err
	}
	// B6
	headerB6 := []interface{}{
		"總流水",
	}
	if err := xlsx.SetSheetRow(sheetName, "B6", &headerB6); err != nil {
		return nil, err
	}
	// B7
	headerB7 := []interface{}{
		"淨盈利(B)",
	}
	if err := xlsx.SetSheetRow(sheetName, "B7", &headerB7); err != nil {
		return nil, err
	}
	row := []interface{}{}
	row2 := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.InternalChargeTotalAmount)
		row2 = append(row2, report.InternalChargeNetProfit)
	}

	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 6, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(6), &row); err != nil {
		return nil, err
	}

	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 7, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(7), &row2); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) WfInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A9合并A10
	headerA9 := []interface{}{
		"下發",
	}
	if err := xlsx.SetSheetRow(sheetName, "A9", &headerA9); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A9", "A10"); err != nil {
		return nil, err
	}
	// B9
	headerB9 := []interface{}{
		"總流水",
	}
	if err := xlsx.SetSheetRow(sheetName, "B9", &headerB9); err != nil {
		return nil, err
	}
	// B10
	headerB10 := []interface{}{
		"淨盈利(C)",
	}
	if err := xlsx.SetSheetRow(sheetName, "B10", &headerB10); err != nil {
		return nil, err
	}
	row := []interface{}{}
	row2 := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.WithdrawTotalAmount)
		row2 = append(row2, report.WithdrawNetProfit)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 9, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(9), &row); err != nil {
		return nil, err
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 10, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(10), &row2); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) DfInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A11合併A12
	headerA11 := []interface{}{
		"代付",
	}
	if err := xlsx.SetSheetRow(sheetName, "A11", &headerA11); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A11", "A12"); err != nil {
		return nil, err
	}
	// B11
	headerB11 := []interface{}{
		"總流水",
	}
	if err := xlsx.SetSheetRow(sheetName, "B11", &headerB11); err != nil {
		return nil, err
	}
	// B12
	headerB12 := []interface{}{
		"淨盈利(D)",
	}
	if err := xlsx.SetSheetRow(sheetName, "B12", &headerB12); err != nil {
		return nil, err
	}
	row := []interface{}{}
	row2 := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.ProxyPayTotalAmount)
		row2 = append(row2, report.ProxyPayNetProfit)

	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 11, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(11), &row); err != nil {
		return nil, err
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 12, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(12), &row2); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) MonthInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A3空白
	headerA3 := []interface{}{
		"",
	}
	if err := xlsx.SetSheetRow(sheetName, "A3", &headerA3); err != nil {
		return nil, err
	}
	// B3
	headerB3 := []interface{}{
		"月份",
	}
	if err := xlsx.SetSheetRow(sheetName, "B3", &headerB3); err != nil {
		return nil, err
	}

	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.Month)
	}

	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(3), &row); err != nil {
		return nil, err
	}

	return xlsx, nil
}

func (l *IncomReportExcelLogic) ReceivedTotalNetProfitInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A8合并B8
	headerA8 := []interface{}{
		"收款總收益(A+B)",
	}
	if err := xlsx.SetSheetRow(sheetName, "A8", &headerA8); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A8", "B8"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.ReceivedTotalNetProfit)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 8, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(8), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) RemitTotalNetProfitInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A14合併B14
	headerA14 := []interface{}{
		"出款總收益(C+D)",
	}
	if err := xlsx.SetSheetRow(sheetName, "A14", &headerA14); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A14", "B14"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.RemitTotalNetProfit)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 14, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(14), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) TotalNetProfitInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A15合併B15
	headerA15 := []interface{}{
		"淨盈利總計",
	}
	if err := xlsx.SetSheetRow(sheetName, "A15", &headerA15); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A15", "B15"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.TotalNetProfit)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 15, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(15), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) CommissionTotalAmountInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A16合併B16
	headerA16 := []interface{}{
		"代理佣金總計",
	}
	if err := xlsx.SetSheetRow(sheetName, "A16", &headerA16); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A16", "B16"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.CommissionTotalAmount)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 16, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(16), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) ProfitGrownRateInfo(xlsx *excelize.File, sheetName string, incomReport []types.IncomReport, rowHeight float64) (*excelize.File, error) {
	// A17合併B17
	headerA17 := []interface{}{
		"盈利成長率(%)",
	}
	if err := xlsx.SetSheetRow(sheetName, "A17", &headerA17); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(sheetName, "A17", "B17"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range incomReport {
		row = append(row, report.ProfitGrowthRate)
	}
	// 設置row 高度
	xlsx.SetRowHeight(sheetName, 17, rowHeight)
	// 塞資料
	if err := xlsx.SetSheetRow(sheetName, "C"+strconv.Itoa(17), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}

func (l *IncomReportExcelLogic) TitleInfo(xlsx *excelize.File, name string, report []types.IncomReport, height float64) (*excelize.File, error) {
	firstMonth := report[0]
	lastMonth := report[len(report)-1]
	firstMonthStrings := strings.Split(firstMonth.Month, "-")
	lastMonthStrings := strings.Split(lastMonth.Month, "-")

	title := []interface{}{
		"CoPo " + firstMonthStrings[0] + "年" + firstMonthStrings[1] + "月~" + lastMonthStrings[0] + "年" + lastMonthStrings[1] + "月 收益报表",
	}

	dataEndColumn, _ := excelize.ColumnNumberToName(1 + len(report) + 1)

	style, _ := xlsx.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 14,
		},
	})
	xlsx.SetCellStyle(name, "A1", "A1", style)

	if err := xlsx.SetSheetRow(name, "A1", &title); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(name, "A1", dataEndColumn+"1"); err != nil {
		return nil, err
	}
	// 設置row 高度
	xlsx.SetRowHeight(name, 1, height)

	return xlsx, nil
}

func (l *IncomReportExcelLogic) AllocInfo(xlsx *excelize.File, name string, report []types.IncomReport, height float64) (*excelize.File, error) {
	// A13合併B13
	headerA13 := []interface{}{
		"撥款總手續費",
	}
	if err := xlsx.SetSheetRow(name, "A13", &headerA13); err != nil {
		return nil, err
	}
	if err := xlsx.MergeCell(name, "A13", "B13"); err != nil {
		return nil, err
	}
	row := []interface{}{}
	for _, report := range report {
		row = append(row, report.TotalAllocHandlingFee)
	}
	// 設置row 高度
	xlsx.SetRowHeight(name, 13, height)
	// 塞資料
	if err := xlsx.SetSheetRow(name, "C"+strconv.Itoa(13), &row); err != nil {
		return nil, err
	}
	return xlsx, nil
}
