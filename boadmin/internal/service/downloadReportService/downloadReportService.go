package downloadReportService

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/aliyunOssService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"time"
)

func CreateDownloadTask(myDB *gorm.DB, ctx context.Context, createDownloadTask *types.CreateDownloadTask) (downReportID int64, fileName string, err error) {

	var isAd string
	var missionName string

	randoms := utils.GetRandomString(5, utils.ALL, utils.MIX)
	fileName = createDownloadTask.Prefix + randoms + time.Now().Format("20060102150405") + ".xlsx"

	if createDownloadTask.IsAdmin {
		isAd = "1"
	} else {
		isAd = "2"
	}

	time.LoadLocation("Asia/Chongqing")
	st, _ := time.Parse("2006-01-02 15:04:05", createDownloadTask.StartAt)
	et, _ := time.Parse("2006-01-02 15:04:05", createDownloadTask.EndAt)
	stf := time.Unix(st.Unix(), 0).Format("20060102150405")
	etf := time.Unix(et.Unix(), 0).Format("20060102150405")
	missionName = createDownloadTask.Prefix + "(" + createDownloadTask.CurrencyCode + ") " + createDownloadTask.Infix + stf + "~" + etf + createDownloadTask.Suffix

	var downReport types.DownloadReportCreate
	downReport.MerchantCode = createDownloadTask.MerchantCode
	downReport.UserId = createDownloadTask.UserId
	downReport.IsAdmin = isAd
	downReport.Status = constants.PROCESSING
	downReport.Type = createDownloadTask.Type
	downReport.FileName = fileName
	downReport.ReqParam = createDownloadTask.ReqParam
	downReport.MissionName = missionName

	if err = myDB.Table("rp_down_report").Create(&downReport).Error; err != nil {
		return 0, "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	downReportID = downReport.ID
	return
}

func UpdateDownloadTask(svcCtx *svc.ServiceContext, ctx context.Context, fileName string, downReportID int64, xlsx *excelize.File, err error) {
	if err != nil {
		if err := model.NewDownReport(svcCtx.MyDB).UpdateFailReport(downReportID); err != nil {
			logx.WithContext(ctx).Error("更新rp_down_report为失败状态，失败。 Err: " + err.Error())
		}
		logx.WithContext(ctx).Error("收款记录产生excel，失败。 Err: " + err.Error())
	} else {
		bf, _ := xlsx.WriteToBuffer()
		errOss := aliyunOssService.UploadFile(ctx, svcCtx, fileName, bf)
		if errOss != nil {
			if err := model.NewDownReport(svcCtx.MyDB).UpdateFailReport(downReportID); err != nil {
				logx.WithContext(ctx).Error("更新rp_down_report为失败状态，失败。 Err: " + err.Error())
			}
			logx.WithContext(ctx).Error("收款记录上传阿里云，失败。 Err: " + errOss.Error())
		} else {
			ss, err := aliyunOssService.DownloadURL(ctx, svcCtx, fileName)
			if err != nil {
				if err := model.NewDownReport(svcCtx.MyDB).UpdateFailReport(downReportID); err != nil {
					logx.WithContext(ctx).Error("更新rp_down_report为失败状态，失败。 Err: " + err.Error())
				}
				logx.WithContext(ctx).Error("取得收款记录下载url，失败。 Err: " + err.Error())
			} else {
				if err := model.NewDownReport(svcCtx.MyDB).UpdateFinishReport(downReportID, ss); err != nil {
					logx.WithContext(ctx).Error("更新rp_down_report为成功状态，失败。 Err: " + err.Error())
				}
			}
		}
	}
}
