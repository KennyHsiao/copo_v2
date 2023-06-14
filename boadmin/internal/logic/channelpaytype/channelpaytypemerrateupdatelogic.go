package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantsRateRecordService"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantChannelRateConfigureLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantChannelRateConfigureLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantChannelRateConfigureLogic {
	return MerchantChannelRateConfigureLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type ChannelPayTypeMerRateUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeMerRateUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeMerRateUpdateLogic {
	return ChannelPayTypeMerRateUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeMerRateUpdateLogic) ChannelPayTypeMerRateUpdate(req *types.MerchantRateUpdateRequest) (err error) {
	var channelData *types.ChannelData
	merChannelRate := types.MerchantChannelRateConfigureRequest{}
	if err = l.svcCtx.MyDB.Table("mc_merchant_channel_rate").Where("id = ?", req.ID).Find(&merChannelRate).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	checkMerChnRate := types.MerchantChannelRateConfigureRequest{}
	copier.Copy(&checkMerChnRate, &req)

	// 取得渠道
	if err = l.svcCtx.MyDB.Table("ch_channels").Where("code = ? ", req.ChannelCode).Take(&channelData).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	checkMerChnRate.CurrencyCode = channelData.CurrencyCode

	err = l.MerchantChannelRateConfigure(checkMerChnRate)

	if err != nil {
		return err
	} else {
		// 发送渠道修改通知
		go func() {
			merchantsService.ChannelChangeNotify(l.svcCtx.MyDB, l.ctx, l.svcCtx, checkMerChnRate.CurrencyCode)
		}()
	}
	return

}

func (l *ChannelPayTypeMerRateUpdateLogic) MerchantChannelRateConfigure(req types.MerchantChannelRateConfigureRequest) (err error) {

	//if req.PayTypeCode == "DF" {
	//	// 有代付單未計算利潤不可異動費率,所以不能轉移代理
	//	if isHas, err := model.NewOrder(l.svcCtx.MyDB).IsHasNotCalculateProfit_DF(req.ChannelPayTypesCode); isHas {
	//		return errorz.New(response.ORDER_NOT_CALCULATED_PROFIT_PLEASE_WAIT, "")
	//	} else if err != nil {
	//		return errorz.New(response.SYSTEM_ERROR, err.Error())
	//	}
	//} else {
	//	// 有支付單未計算利潤不可異動費率,所以不能轉移代理
	//	if isHas, err := model.NewOrder(l.svcCtx.MyDB).IsHasNotCalculateProfit_ZF(req.ChannelPayTypesCode); isHas {
	//		return errorz.New(response.ORDER_NOT_CALCULATED_PROFIT_PLEASE_WAIT, "")
	//	} else if err != nil {
	//		return errorz.New(response.SYSTEM_ERROR, err.Error())
	//	}
	//}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {

		var merchant *types.Merchant
		var designationNo int

		if req.ChannelPayTypesCode != req.ChannelCode+req.PayTypeCode {
			return errorz.New(response.UPDATE_PAYTYPE_NUM_ERROR, "支付類型編碼格式錯誤")
		}

		// 取得商戶
		if merchant, err = model.NewMerchant(db).GetMerchantByCode(req.MerchantCode); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 若商戶多指定時 有指定的話
		if merchant.BillLadingType == "1" && req.Designation == "1" {
			// 代碼必填
			if req.DesignationNo == "" {
				return errorz.New(response.SETTING_MERCHANT_PAY_TYPE_SUB_CODING_NULL_ERROR, "支付方式指定代码不可为空值")
			}
			// 代码格式 1~6
			if designationNo, err = strconv.Atoi(req.DesignationNo); err != nil || designationNo > 6 || designationNo < 1 {
				return errorz.New(response.SETTING_MERCHANT_PAY_TYPE_SUB_CODING_ERROR, "支付方式指定代码格式错误，请输入数字1-6")
			}
		}
		// 檢查費率
		if err = verificationFee(db, req, merchant); err != nil {
			return err
		}

		// 新增或編輯 商戶渠道費率
		if err = l.configure(db, req, merchant); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *ChannelPayTypeMerRateUpdateLogic) configure(db *gorm.DB, req types.MerchantChannelRateConfigureRequest, merchant *types.Merchant) (err error) {
	updatedBy := l.ctx.Value("account").(string)
	var merchantChannelRate *types.MerchantChannelRate

	merchantChannelRateConfigure := types.MerchantChannelRateConfigure{
		MerchantChannelRateConfigureRequest: req,
	}

	if merchant.BillLadingType == "0" { //  若商戶為單指定時
		if merchantChannelRateConfigure.Designation == "1" { // 若本次為指定渠道
			// 若輪巡關閉時 且 本次為指定渠道 則 同一支付類型只能指定一個渠道支付 (把其他渠道支付改為非指定並去除代碼)
			if err = db.Table("mc_merchant_channel_rate").
				Where("merchant_code = ?", req.MerchantCode).
				Where("pay_type_code = ? ", req.PayTypeCode).
				Where("currency_code = ? ", req.CurrencyCode).
				Updates(map[string]interface{}{"designation": "0", "designation_no": ""}).Error; err != nil {
				return
			}
		}
	} else if merchant.BillLadingType == "1" && req.DesignationNo != "" {
		// 若商戶為多指定時開啟時 則 同一支付類型代碼不可重複 (把重複代碼的改為非指定 並去除代碼)
		if err = db.Table("mc_merchant_channel_rate").
			Where("merchant_code = ?", req.MerchantCode).
			Where("pay_type_code = ? ", req.PayTypeCode).
			Where("designation_no = ? ", req.DesignationNo).
			Where("currency_code = ? ", req.CurrencyCode).
			Updates(map[string]interface{}{"designation": "0", "designation_no": ""}).Error; err != nil {
			return
		}
	}

	if merchantChannelRate, err = model.NewMerchantChannelRate(db).GetByMerchantCodeAndChannelPayTypeCode(req.MerchantCode, req.ChannelPayTypesCode); errors.Is(err, gorm.ErrRecordNotFound) {
		// 新增
		merchantChannelRateConfigure.Status = "1"
		err = db.Select("MerchantCode", "ChannelPayTypesCode", "ChannelCode", "PayTypeCode", "Designation", "DesignationNo", "Fee", "HandlingFee", "CurrencyCode").
			Table("mc_merchant_channel_rate").Create(&merchantChannelRateConfigure).Error

		err = merchantsRateRecordService.CreateMerchantRateRecord(db, &types.MerchantRateRecordCreateRequest{
			MerchantCode:       merchantChannelRateConfigure.MerchantCode,
			ChannelPayTypeCode: merchantChannelRateConfigure.ChannelPayTypesCode,
			ModifyType:         "1",
			//BeforeRate:         ,
			AfterRate: merchantChannelRateConfigure.Fee,
			CreatedBy: updatedBy,
		})

		err = merchantsRateRecordService.CreateMerchantRateRecord(db, &types.MerchantRateRecordCreateRequest{
			MerchantCode:       merchantChannelRateConfigure.MerchantCode,
			ChannelPayTypeCode: merchantChannelRateConfigure.ChannelPayTypesCode,
			ModifyType:         "2",
			//BeforeRate:         ,
			AfterRate: merchantChannelRateConfigure.HandlingFee,
			CreatedBy: updatedBy,
		})
		return
	} else if err != nil {
		return
	} else {
		// 編輯
		merchantChannelRateConfigure.ID = merchantChannelRate.ID
		if err = db.Select("MerchantCode", "ChannelPayTypesCode", "ChannelCode", "PayTypeCode", "Designation", "DesignationNo", "Fee", "HandlingFee", "CurrencyCode").
			Table("mc_merchant_channel_rate").Updates(&merchantChannelRateConfigure).Error; err != nil {
			return err
		}

		if merchantChannelRate.Fee != merchantChannelRateConfigure.Fee {
			err = merchantsRateRecordService.CreateMerchantRateRecord(db, &types.MerchantRateRecordCreateRequest{
				MerchantCode:       merchantChannelRateConfigure.MerchantCode,
				ChannelPayTypeCode: merchantChannelRateConfigure.ChannelPayTypesCode,
				ModifyType:         "1",
				BeforeRate:         merchantChannelRate.Fee,
				AfterRate:          merchantChannelRateConfigure.Fee,
				CreatedBy:          updatedBy,
			})
		}

		if merchantChannelRate.HandlingFee != merchantChannelRateConfigure.HandlingFee {
			err = merchantsRateRecordService.CreateMerchantRateRecord(db, &types.MerchantRateRecordCreateRequest{
				MerchantCode:       merchantChannelRateConfigure.MerchantCode,
				ChannelPayTypeCode: merchantChannelRateConfigure.ChannelPayTypesCode,
				ModifyType:         "2",
				BeforeRate:         merchantChannelRate.HandlingFee,
				AfterRate:          merchantChannelRateConfigure.HandlingFee,
				CreatedBy:          updatedBy,
			})
		}

		return err
	}
}
func verificationFee(db *gorm.DB, req types.MerchantChannelRateConfigureRequest, merchant *types.Merchant) (err error) {
	var parentMerchantChannelRate *types.MerchantChannelRate
	var channelPayType *types.ChannelPayType
	var subAgentMerchants []types.Merchant

	// 若有上層代理
	if merchant.AgentParentCode != "" {
		// 要检查上层代理是否有配此渠道
		if parentMerchantChannelRate, err = model.NewMerchantChannelRate(db).GetByMerchantCodeAndChannelPayTypeCode(merchant.AgentParentCode, req.ChannelPayTypesCode); errors.Is(err, gorm.ErrRecordNotFound) {
			return errorz.New(response.SETTING_NO_PARENT_MERCHANT_RATE_ERROR, "代理商户父层费率未设置，请先设置父层商户费率")
		}
		// 檢核費率為開啟 要檢查代理費率  (若是商户使用此API 就一定要检查)
		if merchant.RateCheck == "1" {
			if err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			} else if req.Fee < parentMerchantChannelRate.Fee {
				return errorz.New(response.SETTING_MERCHANT_RATE_LOWER_PARENT_ERROR, "配置商户费率不可小于父层费率")
			} else if req.HandlingFee < parentMerchantChannelRate.HandlingFee {
				return errorz.New(response.SETTING_MERCHANT_CHAREG_LOWER_PARENT_ERROR, "配置商户手续费不可小于父层手续费")
			}
		}
	}

	// 檢核費率為開啟需 檢察系統費率
	if channelPayType, err = model.NewChannelPayType(db).GetByCode(req.ChannelPayTypesCode); err != nil {
		return errorz.New(response.PAY_TYPE_NOT_EXIST, err.Error())
	} else if req.Fee < channelPayType.Fee && (merchant.RateCheck == "1") {
		return errorz.New(response.SETTING_MERCHANT_RATE_MIN_CHARGE_ERROR, "配置商户费率低消不可低于渠道成本费率低消")
	} else if req.HandlingFee < channelPayType.HandlingFee && (merchant.RateCheck == "1") {
		return errorz.New(response.SETTING_MERCHANT_CHARGE_ERROR, "配置商户手续费不可低于渠道成本手续费")
	}

	// 渠道单笔限额最小值*費率+手续费 > 渠道单笔限额最小值 (表示手續費會超過訂單金額)
	if utils.FloatAdd(utils.FloatDiv(channelPayType.SingleMinCharge, 100), req.Fee)+req.HandlingFee > channelPayType.SingleMinCharge {
		return errorz.New(response.SETTING_MERCAHNT_INCOME_OVER_CHN_MIN_LIMIT, "配置费率和手续费计算高于渠道单笔限额最小值")
	}

	if merchant.AgentLayerCode != "" {
		// 取得下級商戶
		if subAgentMerchants, err = model.NewMerchant(db).GetDescendantAgents(merchant.AgentLayerCode, false); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 若有下級商戶 則需要檢查下級商戶費率
		for _, subAgentMerchant := range subAgentMerchants {
			if subAgentMerchant.RateCheck == "0" {
				//(系统管理员操作) 下層商戶如果不需檢核費率 則跳過
				continue
			}
			var subMerchantChannelRate *types.MerchantChannelRate
			if subMerchantChannelRate, err = model.NewMerchantChannelRate(db).GetByMerchantCodeAndChannelPayTypeCode(subAgentMerchant.Code, req.ChannelPayTypesCode); errors.Is(err, gorm.ErrRecordNotFound) {
				err = nil
			} else if err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			} else if req.Fee > subMerchantChannelRate.Fee {
				return errorz.New(response.SETTING_MERCHANT_RATE_OVER_LOWER_MERCHANT_ERROR, "配置商户费率不得高于下层商户费率")
			} else if req.HandlingFee > subMerchantChannelRate.HandlingFee {
				return errorz.New(response.SETTING_MERCHANT_CHARGE_OVER_LOWER_MERCHANT_ERROR, "配置商户手续费不得高于下层商户手续费")
			}
		}
	}

	return
}

type respVO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	traceId string `json:"traceId"`
}
