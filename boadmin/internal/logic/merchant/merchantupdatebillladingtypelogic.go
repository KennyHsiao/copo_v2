package merchant

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateBillLadingTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateBillLadingTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateBillLadingTypeLogic {
	return MerchantUpdateBillLadingTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateBillLadingTypeLogic) MerchantUpdateBillLadingType(req types.MerchantUpdateBillLadingTypeRequest) (err error) {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		var merchant types.Merchant

		// 取得商戶
		if err = db.Table("mc_merchants").Where("code = ?", req.Code).Take(&merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 狀態沒變結束
		if merchant.BillLadingType == req.BillLadingType {
			return
		}

		if req.BillLadingType == "1" {
			if err = openBillLadingType(db, merchant.Code); err != nil {
				return
			}

		} else {
			if err = closeBillLadingType(db, merchant.Code); err != nil {
				return
			}
		}

		if err = udpateBillLadingType(db, req.Code, req.BillLadingType); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})

}

func openBillLadingType(db *gorm.DB, merchantCode string) (err error) {
	//開 => 將商戶渠道費率的已指定的渠道 代碼賦值 "1"
	if err = db.Table("mc_merchant_channel_rate").
		Where("merchant_code = ?", merchantCode).
		Where("designation = ? ", "1").
		Updates(map[string]interface{}{"designation_no": "1"}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func closeBillLadingType(db *gorm.DB, merchantCode string) (err error) {
	//關 => 此商戶同個支付類型中 全部已指定的清掉代碼
	//1.清掉代碼
	//2.保留代碼最小的指定狀態 其餘改為未指定
	var merchantChannelRates []types.MerchantChannelRate
	keepIDs := []int64{0}
	payTypeRateMap := make(map[string]types.MerchantChannelRate)

	// 取得此商戶全部已指定費率
	if err = db.Table("mc_merchant_channel_rate").
		Where("merchant_code = ?", merchantCode).
		Where("designation = ? ", "1").
		Find(&merchantChannelRates).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	// 紀錄每種PayTypeCode下的最小指定編號
	for _, rate := range merchantChannelRates {
		mapRate, isExist := payTypeRateMap[rate.PayTypeCode]
		if !isExist || rate.DesignationNo < mapRate.DesignationNo {
			payTypeRateMap[rate.PayTypeCode] = rate
		}
	}

	for _, rate := range payTypeRateMap {
		keepIDs = append(keepIDs, rate.ID)
	}

	// 將保留的清掉代碼
	if err = db.Table("mc_merchant_channel_rate").
		Where("merchant_code = ?", merchantCode).
		Where("id in ? ", keepIDs).
		Updates(map[string]interface{}{"designation_no": ""}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	// 將不保留的 清掉代碼並改為非指定
	if err = db.Table("mc_merchant_channel_rate").
		Where("merchant_code = ?", merchantCode).
		Where("id not in ? ", keepIDs).
		Updates(map[string]interface{}{"designation_no": "", "designation": "0"}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func udpateBillLadingType(db *gorm.DB, code string, billLadingType string) error {
	return db.Table("mc_merchants").Where("code = ?", code).
		Updates(map[string]interface{}{"bill_lading_type": billLadingType}).Error
}
