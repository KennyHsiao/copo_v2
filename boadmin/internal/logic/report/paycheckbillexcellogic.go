package report

import (
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/common/excelizeutil"
	"context"
	"github.com/xuri/excelize/v2"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayCheckBillExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayCheckBillExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayCheckBillExcelLogic {
	return PayCheckBillExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayCheckBillExcelLogic) PayCheckBillExcel(req *types.PayCheckBillQueryRequestX) (xlsx *excelize.File, err error) {

	sheetName := "支付对帐报表"
	var rowHeight float64 = 17

	//渠得資料
	var resp *types.PayCheckBillQueryResponse
	if resp, err = reportService.PayCheckBill(l.svcCtx.MyDB, req, l.ctx); err != nil {
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
		"商户编号", "渠道名称",
		"支付类型", "支付笔数",
		"支付总金额", "总手续费",
		"支付手续费", "代理佣金",
		"我司佣金", "调整金额",
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {
		row := []interface{}{
			item.MerchantCode, item.ChannelName,
			item.PayTypeName, item.TotalNum,
			item.TotalOrderAmount, item.TotalHandlingFee,
			item.PayTotalTransferHandlingFee, item.AgentCommission,
			item.SystemCommission, item.AdjustmentAmount,
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
		"", resp.TotalNum,
		resp.TotalOrderAmount, resp.TotalHandlingFee,
		resp.PayTotalTransferHandlingFee, resp.AgentCommission,
		resp.SystemCommission, resp.AdjustmentAmount,
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &totalRow); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "J1", headerStyle)
	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "J"+footerRow, footerStyle)

	return
}
