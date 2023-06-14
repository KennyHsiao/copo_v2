package merchantchannelrate

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantRateResetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantRateResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantRateResetLogic {
	return MerchantRateResetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantRateResetLogic) MerchantRateReset(req *types.MerchantRateResetRequest) (err error) {

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {

		var merchant *types.Merchant
		var merchants []types.Merchant
		merchantModel := model.NewMerchant(db)
		merchantRateModel := model.NewMerchantChannelRate(db)
		// 取得商戶
		if merchant, err = merchantModel.GetMerchantByCode(req.MerchantCode); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 删除自身费率
		if err = merchantRateModel.DeleteByMerchantCodeAndChannelPayTypeCode(req.MerchantCode, req.ChannelPayTypesCode); err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		// 判断是否有层级编号
		if len(merchant.AgentLayerCode) > 0 {
			// 取得所有子孙代理
			if merchants, err = merchantModel.GetDescendantAgents(merchant.AgentLayerCode, false); err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			// 子孙代理的费率一并删除
			for _, mer := range merchants {
				if err = merchantRateModel.DeleteByMerchantCodeAndChannelPayTypeCode(mer.Code, req.ChannelPayTypesCode); err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			}
		}

		return
	})
}
