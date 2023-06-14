package merchant

import (
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateStatusLogic {
	return MerchantUpdateStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateStatusLogic) MerchantUpdateStatus(req types.MerchantUpdateStatusRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		var merchant types.Merchant

		if err := db.Table("mc_merchants").Where("code = ?", req.Code).Take(&merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		switch req.Status {
		case constants.MerchantStatusEnable:
			var parentMerchant types.Merchant
			if merchant.AgentParentCode != "" {
				if err := db.Table("mc_merchants").Where("code = ?", merchant.AgentParentCode).Take(&parentMerchant).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
				if parentMerchant.AgentStatus == constants.MerchantAgentStatusDisable {
					return errorz.New(response.UPPER_LAYER_STATUS_NOT_OPEN, "上层相关代理角色或状态尚未启用，请确认")
				}
			}
		case constants.MerchantStatusDisable, constants.MerchantStatusClear:
			if merchant.AgentStatus == constants.MerchantAgentStatusEnable {
				return errorz.New(response.PLEASE_DISABLE_LAYER_STATUS, "请先禁用代理身份，再进行启用状态操作")
			}
		}

		if err := udpateStatus(db, req.Code, req.Status); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func udpateStatus(db *gorm.DB, code string, status string) error {
	return db.Table("mc_merchants").Where("code = ?", code).
		Updates(map[string]interface{}{"status": status}).Error
}
