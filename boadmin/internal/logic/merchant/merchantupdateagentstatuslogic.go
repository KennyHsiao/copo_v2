package merchant

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUpdateAgentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUpdateAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUpdateAgentStatusLogic {
	return MerchantUpdateAgentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUpdateAgentStatusLogic) MerchantUpdateAgentStatus(req types.MerchantUpdateAgentStatusRequest) error {

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		var merchant types.Merchant

		if err := db.Table("mc_merchants").Preload("Users").Where("code = ?", req.Code).Take(&merchant).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if req.AgentStatus == constants.MerchantAgentStatusDisable && len(merchant.AgentLayerCode) > 0 {
			//代理狀態禁用時 子代理都要啟用狀態都要改成禁用
			if err := db.Table("mc_merchants").
				Where("agent_layer_code LIKE ?", merchant.AgentLayerCode+"%").
				Where("agent_layer_code != ?", merchant.AgentLayerCode).
				Update("agent_status", constants.MerchantAgentStatusDisable).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			//全部子孫代理主帳號都要恢復為普通商戶角色
			ux := model.NewMerchant(db)
			merchants, err := ux.GetDescendantAgents(merchant.AgentLayerCode, true)
			if err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			for _, mer := range merchants {
				for _, user := range mer.Users {
					if user.DisableDelete == "1" {
						if err = db.Table("au_user_roles").Where("user_id = ?", user.ID).
							Updates(map[string]interface{}{"role_id": 2}).Error; err != nil {
							return errorz.New(response.DATABASE_FAILURE, err.Error())
						}
					}
				}
			}
		}

		if req.AgentStatus == constants.MerchantAgentStatusEnable && len(merchant.AgentLayerCode) == 0 {
			//代理狀態啟用時 若沒代理編號則要給最新總代編號
			merchant.AgentLayerCode = model.NewMerchant(db).GetNextGeneralAgentCode()
		}
		if req.AgentStatus == constants.MerchantAgentStatusEnable {
			//代理狀態啟用時 需將主帳號改為代理角色
			for _, user := range merchant.Users {
				if user.DisableDelete == "1" {
					if err := db.Table("au_user_roles").Where("user_id = ?", user.ID).
						Updates(map[string]interface{}{"role_id": 3}).Error; err != nil {
						return errorz.New(response.DATABASE_FAILURE, err.Error())
					}
				}
			}
		}

		merchant.AgentStatus = req.AgentStatus
		if err := db.Table("mc_merchants").Updates(types.MerchantUpdate2{
			Merchant: merchant,
		}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})

}

func (l *MerchantUpdateAgentStatusLogic) ChangeMerAccountRole(merchant, roleId string) {

}
