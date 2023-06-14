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

type AppropriationCheckBillExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppropriationCheckBillExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) AppropriationCheckBillExcelLogic {
	return AppropriationCheckBillExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppropriationCheckBillExcelLogic) AppropriationCheckBillExcel(req *types.AppropriationCheckBillQueryRequest) (xlsx *excelize.File, err error) {
	sheetName := "拨款对帐报表"
	var rowHeight float64 = 17

	//取得資料
	var resp *types.AppropriationCheckBillQueryResponse
	if resp, err = reportService.AppropriationCheckBill(l.svcCtx.MyDB, req, l.ctx); err != nil {
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
		"渠道名称", "拨款笔数",
		"拨款总金额", "拨款手续费",
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)
	// 迴圈建置資料
	for i, item := range resp.List {
		row := []interface{}{
			item.ChannelName, item.AppropriationCount,
			item.AppropriationAmount, item.AppropriationHandlingFee,
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
		"加总", resp.TotalCount,
		resp.TotalAmount, resp.TotalHandlingFee,
	}
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &totalRow); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "D1", headerStyle)
	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "D"+footerRow, footerStyle)

	return
}
