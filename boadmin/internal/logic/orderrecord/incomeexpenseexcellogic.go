package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/service/downloadReportService"
	orderrecordService "com.copo/bo_service/boadmin/internal/service/orderrecord"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/excelizeutil"
	"context"
	"encoding/json"
	"github.com/xuri/excelize/v2"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncomeExpenseExcelAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncomeExpenseExcelAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) IncomeExpenseExcelAllLogic {
	return IncomeExpenseExcelAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncomeExpenseExcelAllLogic) IncomeExpenseExcelAll(req *types.IncomeExpenseQueryRequestX) (err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	userId := l.ctx.Value("userId").(json.Number)
	userIdint, _ := userId.Int64()
	isAdmin := l.ctx.Value("isAdmin").(bool)
	requestBytes, _ := json.Marshal(req)

	createDownloadTask := &types.CreateDownloadTask{
		Prefix:       "收支纪录数据",
		Infix:        "",
		Suffix:       "",
		IsAdmin:      isAdmin,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		CurrencyCode: req.CurrencyCode,
		MerchantCode: merchantCode,
		UserId:       userIdint,
		ReqParam:     string(requestBytes),
		Type:         constants.INCOME_EXPENSE_REPORT,
	}
	downReportID, fileName, err := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
	if err != nil {
		return err
	}
	go func() {
		xlsx, err := l.DoIncomeExpenseExcelAll(req)
		if err != nil {
			logx.WithContext(l.ctx).Error("DoIncomeExpenseExcelAll Err: " + err.Error())
		}
		downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
	}()

	return
}

func (l *IncomeExpenseExcelAllLogic) DoIncomeExpenseExcelAll(req *types.IncomeExpenseQueryRequestX) (xlsx *excelize.File, err error) {

	sheetName := "收支记录(后台)"
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.IncomeExpenseQueryResponseX
	if resp, err = orderrecordService.IncomeExpenseQueryAll(l.svcCtx.MyDB, *req, nil); err != nil {
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
		"商户号", "平台订单号",
		"商户订单号", "渠道名称",
		"交易类型", "支付类型",
		"余额类型", "变动前金额",
		"变动金额", "变动后金额",
		"备注", "交易时间",
	}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {

		row := []interface{}{
			item.MerchantCode, item.OrderNo,
			item.MerchantOrderNo, item.ChannelName,
			excelizeutil.GetBalanceRecordTransactionTypeName(item.TransactionType), item.PayTypeName,
			excelizeutil.GetBalanceType(item.BalanceType), item.BeforeBalance,
			item.TransferAmount, item.AfterBalance,
			item.Comment, item.CreatedAt.FormatAndZero("2006-01-02 15:04:05", "Asia/Taipei"),
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
