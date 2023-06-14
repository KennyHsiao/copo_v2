package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"gorm.io/gorm"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantTransferParentAgentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantTransferParentAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantTransferParentAgentLogic {
	return MerchantTransferParentAgentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantTransferParentAgentLogic) MerchantTransferParentAgent(req types.MerchantTransferParentAgentRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		var merchant types.Merchant
		var parentMerchant types.Merchant

		// 取得商戶
		if merchant, err = getMerchant(db, req.Code); err != nil {
			return err
		}
		// 取得上級代理商戶
		if parentMerchant, err = getMerchant(db, req.AgentParentCode); err != nil {
			return err
		}
		// 驗證
		if err = verifyTransfer(db, merchant, parentMerchant); err != nil {
			return err
		}
		// 轉移代理
		var agentLayerCode string
		if agentLayerCode, err = model.NewAgentRecord(db).GetNextAgentLayerCode(req.Code, parentMerchant.Code, parentMerchant.AgentLayerCode); err != nil {
			return err
		}
		if err = setParentAgentLoop(db, agentLayerCode, &merchant, &parentMerchant); err != nil {
			return err
		}

		return nil
	})

}

func getMerchant(db *gorm.DB, code string) (merchant types.Merchant, err error) {
	if err = db.Table("mc_merchants").Where("code = ?", code).Take(&merchant).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return merchant, errorz.New(response.INVALID_MERCHANT_CODING, err.Error())
	} else if err != nil {
		return merchant, errorz.New(response.SETTING_AGENT_ERROR, err.Error())
	}
	return
}

func verifyTransfer(db *gorm.DB, merchant types.Merchant, parentMerchant types.Merchant) error {
	if merchant.AgentParentCode == parentMerchant.Code {
		return errorz.New(response.CHANGE_AGENT_TO_SAME_LAYER_ERROR, "目前已指定於該代理，就不可以重複指定")
	}

	if parentMerchant.AgentStatus == constants.MerchantAgentStatusDisable ||
		parentMerchant.Status == constants.MerchantStatusDisable ||
		parentMerchant.Status == constants.MerchantStatusClear ||
		len(parentMerchant.AgentLayerCode) == 0 {
		return errorz.New(response.PARENT_MERCHANT_IS_NOT_AGENT, "指定的上層代理，目前不可添加商戶或配置代理 = 代理或該代理已被停權)。")
	}

	if len(merchant.AgentLayerCode) > 0 && len(merchant.AgentParentCode) == 0 {
		return errorz.New(response.MERCHANT_IS_LAYER_ONE_CANT_SETTING, "商戶為總代身份，無法轉換至其他代理下")
	}

	// 當 "目標代理編碼" 包含 "自身代理編碼" 則表示轉移代理到自身子孫代理
	if len(merchant.AgentLayerCode) > 0 &&
		len(merchant.AgentLayerCode) < len(parentMerchant.AgentLayerCode) &&
		strings.Contains(parentMerchant.AgentLayerCode, merchant.AgentLayerCode) {
		return errorz.New(response.CANNOT_BE_TRANSFERRED_TO_SUBAGENT, "不可轉移至自身的子代理")
	}

	//// 有代付單未計算利潤不可異動費率,所以不能轉移代理
	//if isHas, err := model.NewOrder(db).IsHasNotCalculateProfit_DF(""); isHas {
	//	return errorz.New(response.ORDER_NOT_CALCULATED_PROFIT_PLEASE_WAIT, "")
	//} else if err != nil {
	//	return errorz.New(response.SYSTEM_ERROR, err.Error())
	//}
	//// 有支付單未計算利潤不可異動費率,所以不能轉移代理
	//if isHas, err := model.NewOrder(db).IsHasNotCalculateProfit_ZF(""); isHas {
	//	return errorz.New(response.ORDER_NOT_CALCULATED_PROFIT_PLEASE_WAIT, "")
	//} else if err != nil {
	//	return errorz.New(response.SYSTEM_ERROR, err.Error())
	//}

	return nil
}

// 設置自身和子代理的代理變號 直到再無後代
func setParentAgentLoop(db *gorm.DB, agentLayerCode string, merchant *types.Merchant, parentMerchant *types.Merchant) (err error) {
	var subAgentMerchants []types.Merchant

	// 變更代理號
	merchant.AgentLayerCode = agentLayerCode
	merchant.AgentParentCode = parentMerchant.Code
	if err = db.Table("mc_merchants").Updates(types.MerchantUpdate2{
		Merchant: *merchant,
	}).Error; err != nil {
		return errorz.New(response.SETTING_AGENT_ERROR, err.Error())
	}

	// 變更代理要清掉全部配置渠道
	if err = db.Where("merchant_code = ?", merchant.Code).Delete(&types.MerchantChannelRate{}).Error; err != nil {
		return errorz.New(response.SETTING_AGENT_ERROR, err.Error())
	}

	// 取得子商戶
	if subAgentMerchants, err = model.NewMerchant(db).GetSubAgents(merchant.Code); err != nil {
		return errorz.New(response.SETTING_AGENT_ERROR, err.Error())
	}
	// 子商戶同樣執行此LOOP
	if len(subAgentMerchants) > 0 {
		for _, subAgentMerchant := range subAgentMerchants {
			// 層級代碼依index遞增
			var subAgentLayerCode string
			if subAgentLayerCode, err = model.NewAgentRecord(db).GetNextAgentLayerCode(subAgentMerchant.Code, merchant.Code, merchant.AgentLayerCode); err != nil {
				return err
			}
			setParentAgentLoop(db, subAgentLayerCode, &subAgentMerchant, merchant)
		}
	}

	return err
}
