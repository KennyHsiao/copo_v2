package allocorder

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

type AllocRecordExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAllocRecordExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) AllocRecordExcelLogic {
	return AllocRecordExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AllocRecordExcelLogic) AllocRecordExcel(req *types.AllocRecordQueryAllRequestX) (xlsx *excelize.File, err error) {

	// 設置i18n
	utils.SetI18n(req.Language)
	sheetName := i18n.Sprintf("Allocation of record")
	var rowHeight float64 = 17

	// 取得资料
	var resp *types.AllocRecordQueryAllResponseX
	var totalInfo *types.AllocRecordTotalInfoResponse
	if resp, err = orderrecordService.AllocRecordQueryAll(l.svcCtx.MyDB, *req); err != nil {
		return
	}
	if totalInfo, err = orderrecordService.AllocRecordTotalInfo(l.svcCtx.MyDB, *req); err != nil {
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
		"平台订单号", "渠道名称",
		"银行", "开户省",
		"开户市", "开户姓名",
		"银行卡号", "出款金额",
		"手續費", "費率",
		"訂單狀態", "備註",
		"提單時間", "交易時間",
		"提单人员",
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

		row := []interface{}{
			orderNo, item.ChannelName,
			item.MerchantBankName, item.MerchantBankProvince,
			item.MerchantBankCity, item.MerchantAccountName,
			item.MerchantBankAccount, item.OrderAmount,
			item.HandlingFee, item.Fee,
			excelizeutil.GetTxOrderStatusName(item.Status), memo,
			createdAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"), item.TransAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
			item.CreatedBy,
		}

		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	//xlsx.SetCellFormula(sheetName, "H"+strconv.Itoa(len(resp.List)+2), "=SUM(H2:"+"H"+strconv.Itoa(len(resp.List)+1)+")")
	//xlsx.SetCellFormula(sheetName, "I"+strconv.Itoa(len(resp.List)+2), "=SUM(I2:"+"I"+strconv.Itoa(len(resp.List)+1)+")")
	//設置總額
	row := []interface{}{
		"总拨款金额 :", totalInfo.TotalOrderAmount,
		"总拨款手续费 :", totalInfo.TotalTransferHandlingFee,
	}

	xlsx.SetRowHeight(sheetName, len(resp.List)+2, rowHeight)
	if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(len(resp.List)+2), &row); err != nil {
		return
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)

	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "O1", headerStyle)

	// 設置頁腳 style
	footerRow := strconv.Itoa(len(resp.List) + 2)
	xlsx.SetCellStyle(sheetName, "A"+footerRow, "O"+footerRow, bottomStyle)

	return
}
