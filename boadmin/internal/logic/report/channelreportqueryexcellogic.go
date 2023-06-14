package report

import (
	"com.copo/bo_service/boadmin/internal/service/downloadReportService"
	reportService "com.copo/bo_service/boadmin/internal/service/report"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"context"
	"encoding/json"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelReportQueryExcelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelReportQueryExcelLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelReportQueryExcelLogic {
	return ChannelReportQueryExcelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelReportQueryExcelLogic) ChannelReportQueryExcel(req *types.ChannelReportQueryRequestX) (err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	userId := l.ctx.Value("userId").(json.Number)
	userIdint, _ := userId.Int64()
	isAdmin := l.ctx.Value("isAdmin").(bool)
	requestBytes, _ := json.Marshal(req)

	createDownloadTask := &types.CreateDownloadTask{
		Prefix:       "渠道报表数据",
		Infix:        "",
		Suffix:       "",
		IsAdmin:      isAdmin,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		CurrencyCode: req.CurrencyCode,
		MerchantCode: merchantCode,
		UserId:       userIdint,
		ReqParam:     string(requestBytes),
		Type:         constants.CHANNEL_REPORT,
	}
	downReportID, fileName, err := downloadReportService.CreateDownloadTask(l.svcCtx.MyDB, l.ctx, createDownloadTask)
	if err != nil {
		return err
	}
	go func() {
		xlsx, err := newGenerateChannelExcel(l.svcCtx.MyDB, req)
		if err != nil {
			logx.WithContext(l.ctx).Error("newGenerateChannelExcel Err: " + err.Error())
		}
		downloadReportService.UpdateDownloadTask(l.svcCtx, l.ctx, fileName, downReportID, xlsx, err)
	}()

	return
}

func newGenerateChannelExcel(db *gorm.DB, req *types.ChannelReportQueryRequestX) (xlsx *excelize.File, err error) {
	sheetName := "渠道报表"
	var rowHeight float64 = 17

	//取得資料
	var channelReportResp *types.ChannelReportQueryresponse
	req.PageSize = 0
	if channelReportResp, err = reportService.InterChannelReport(db, req, nil); err != nil {
		return
	}

	var channelReportTotalResp *types.ChannelReportTotalResponse
	req.PageSize = 0
	if channelReportTotalResp, err = reportService.InterChannelReportTotal(db, req, nil); err != nil {
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
		"渠道名称", "渠道编码",
		"交易类型", "支付类型",
		"订单总数", "订单总额",
		"成功总数", "成功总额",
		"成功率", "渠道费率",
		"渠道手续", "系统成本",
		"系统盈利",
	}
	w.SetRow("A1", header,
		excelize.RowOpts{StyleID: headerStyle, Height: rowHeight, Hidden: false})

	// 迴圈建置資料
	for i, item := range channelReportResp.List {
		row := []interface{}{
			item.ChannelName, item.ChannelCode,
			item.TransactionType, item.PayTypeName,
			item.OrderQuantity, item.OrderAmount,
			item.SuccessQuantity, item.SuccessAmount,
			item.SuccessRate, item.Fee,
			item.HandlingFee, item.SystemCost,
			item.SystemProfit,
		}

		cell, _ := excelize.CoordinatesToCellName(1, i+2)
		w.SetRow(cell, row)
	}

	// 設置總金額
	totalRow := []interface{}{
		excelize.Cell{StyleID: footerStyle, Value: "加总"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: "--"},
		excelize.Cell{StyleID: footerStyle, Value: "--"}, excelize.Cell{StyleID: footerStyle, Value: channelReportTotalResp.TotalCost},
		excelize.Cell{StyleID: footerStyle, Value: channelReportTotalResp.TotalProfit},
	}
	footerRow := strconv.Itoa(len(channelReportResp.List) + 2)
	w.SetRow("A"+footerRow, totalRow)
	w.Flush()

	return
}
