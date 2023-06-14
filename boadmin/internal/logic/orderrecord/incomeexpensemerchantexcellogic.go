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

type IncomeMerchantExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomeMerchantExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomeMerchantExcelLogic {
	return IncomeMerchantExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomeMerchantExcelLogic) IncomeMerchantExcel(req *types.IncomeExpenseQueryRequestX) (xlsx *excelize.File, err error) {

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.MerchantCode = jwtMerchantCode
	}
	// 設置i18n
	utils.SetI18n(req.Language)

	sheetName := i18n.Sprintf("Receipt Payment Report")
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.IncomeExpenseQueryResponseX
	if resp, err = orderrecordService.IncomeExpenseQueryAll(l.svcCtx.MyDB, *req, l.ctx); err != nil {
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

	// 設置標題
	header := []interface{}{
		i18n.Sprintf("Platform order number"), i18n.Sprintf("Merchant order number"),
		i18n.Sprintf("Transaction type"), i18n.Sprintf("Payment types"),
		i18n.Sprintf("Balance type"), i18n.Sprintf("Amount before edit"),
		i18n.Sprintf("Edit amount"), i18n.Sprintf("Amount after edit"),
		i18n.Sprintf("Note"), i18n.Sprintf("Created time"),
		i18n.Sprintf("Currency"),
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {

		row := []interface{}{
			item.OrderNo, item.MerchantOrderNo,
			excelizeutil.GetBalanceRecordTransactionTypeName(item.TransactionType), item.PayTypeName,
			excelizeutil.GetBalanceType(item.BalanceType), item.BeforeBalance,
			item.TransferAmount, item.AfterBalance,
			item.Comment, item.CreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.CurrencyCode,
		}

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		// 把资料塞入 row
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "T1", headerStyle)

	return
}
