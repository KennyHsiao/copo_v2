package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
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

type MerchantExportListViewExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantExportListViewExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantExportListViewExcelLogic {
	return MerchantExportListViewExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantExportListViewExcelLogic) MerchantExportListViewExcel(req types.MerchantQueryListViewRequestX) (xlsx *excelize.File, err error) {

	// 設置i18n
	utils.SetI18n(req.Language)

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
		return l.excelForAgent(req)
	} else {
		return l.excelForSys(req)
	}
}

func getStatusName(status string) string {
	name := ""
	switch status {
	case "0":
		name = i18n.Sprintf("Disable")
	case "1":
		name = i18n.Sprintf("Enable")
	case "2":
		name = i18n.Sprintf("Settle")
	}
	return name
}

func (l *MerchantExportListViewExcelLogic) excelForSys(req types.MerchantQueryListViewRequestX) (xlsx *excelize.File, err error) {

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}

	sheetName := i18n.Sprintf("Merchant list")
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.MerchantQueryListViewResponse
	if resp, err = model.NewMerchantListView(l.svcCtx.MyDB).QueryListView(req); err != nil {
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
		i18n.Sprintf("Merchant id"), i18n.Sprintf("Login Name"), i18n.Sprintf("Level Number"),
		i18n.Sprintf("Withdraw balance"), i18n.Sprintf("Payout balance"),
		i18n.Sprintf("Total frozen amount"), i18n.Sprintf("Total commission"),
		i18n.Sprintf("Merchant status"), i18n.Sprintf("Agent status")}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {
		row := []interface{}{item.Code, item.AccountName, item.AgentLayerCode, item.XfBalance, item.DfBalance, item.FrozenAmount, item.Commission, getStatusName(item.Status), getStatusName(item.AgentStatus)}
		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		// 塞資料
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "Z1", headerStyle)

	return
}

func (l *MerchantExportListViewExcelLogic) excelForAgent(req types.MerchantQueryListViewRequestX) (xlsx *excelize.File, err error) {

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if jwtMerchantCode != "" {
		req.JwtMerchantCode = jwtMerchantCode
	}

	sheetName := i18n.Sprintf("Merchant list")
	var rowHeight float64 = 17

	// 取得資料
	var resp *types.MerchantQueryListViewResponse
	if resp, err = model.NewMerchantListView(l.svcCtx.MyDB).QueryListView(req); err != nil {
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
		i18n.Sprintf("Merchant id"), i18n.Sprintf("Level Number"),
		i18n.Sprintf("Group name"), i18n.Sprintf("Messenging app"),
		i18n.Sprintf("Company name"), i18n.Sprintf("Total available balance"),
		i18n.Sprintf("Total frozen amount"), i18n.Sprintf("Merchant status"),
		i18n.Sprintf("Agent status")}
	if err = xlsx.SetSheetRow(sheetName, "A1", &header); err != nil {
		return
	}
	// 設置標題row 高度
	xlsx.SetRowHeight(sheetName, 1, rowHeight)

	// 迴圈建置資料
	for i, item := range resp.List {
		row := []interface{}{
			item.Code, item.AgentLayerCode,
			item.Contact.GroupName, item.Contact.CommunicationSoftware,
			item.BizInfo.CompanyName, item.XfBalance + item.DfBalance,
			item.FrozenAmount, getStatusName(item.Status),
			getStatusName(item.AgentStatus)}
		// 設置row 高度
		xlsx.SetRowHeight(sheetName, i+2, rowHeight)
		// 塞資料
		if err = xlsx.SetSheetRow(sheetName, "A"+strconv.Itoa(i+2), &row); err != nil {
			return
		}
	}

	// 自動設置col 寬度
	excelizeutil.SetColWidthAuto(xlsx, sheetName)
	// 設置標題 style
	xlsx.SetCellStyle(sheetName, "A1", "Z1", headerStyle)

	return
}
