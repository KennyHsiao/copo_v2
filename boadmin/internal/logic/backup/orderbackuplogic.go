package backup

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type OrderBackUpLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderBackUpLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderBackUpLogic {
	return OrderBackUpLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderBackUpLogic) OrderBackUp() error {
	t := time.Now()
	beforeDate := t.AddDate(0, -3, -1).Format("2006-01-02 15:04:05")
	logx.WithContext(l.ctx).Infof("开始搬移资料，搬移时间'%s'前的资料", beforeDate)

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {

		for i := 0; i <= 200; i++ {
			var orders []types.OrderX
			if err := db.Table("tx_orders").Where("created_at < ?", beforeDate).Limit(5000).Find(&orders).Error; err != nil {
				logx.WithContext(l.ctx).Error("搬移资料错误，查询tx_orders错误 : ", err.Error())
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			if len(orders) > 0 {
				for _, order := range orders {
					if err := db.Table("bu_orders").Create(order).Error; err != nil {
						logx.WithContext(l.ctx).Error("搬移资料错误，新增bu_orders错误 : ", err.Error())
						return errorz.New(response.DATABASE_FAILURE, err.Error())
					}
				}
				if err := db.Table("tx_orders").Delete(&orders).Error; err != nil {
					logx.WithContext(l.ctx).Error("搬移资料错误，删除tx_orders错误 : ", err.Error())
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			} else {
				break
			}

			//if err := db.Table("bu_orders").CreateInBatches(orders, len(orders)).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}

			//if err := db.Table("bu_order_actions").CreateInBatches(orderActions, len(orderActions)).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}
			//
			//if err := db.Table("bu_orders_fee_profit").CreateInBatches(orderFeeProfits, len(orderFeeProfits)).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}
			//
			//if err := db.Table("bu_order_channels").CreateInBatches(orderChannels, len(orderChannels)).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}
		}

		for i := 0; i <= 200; i++ {
			var orderActions []types.OrderActionX
			if err := db.Table("tx_order_actions").Where("created_at < ?", beforeDate).Limit(5000).Find(&orderActions).Error; err != nil {
				logx.WithContext(l.ctx).Error("搬移资料错误，查询tx_order_actions错误 : ", err.Error())
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			if len(orderActions) > 0 {
				for _, order := range orderActions {
					if err := db.Table("bu_order_actions").Create(order).Error; err != nil {
						logx.WithContext(l.ctx).Error("搬移资料错误，新增bu_order_actions错误 : ", err.Error())
						return errorz.New(response.DATABASE_FAILURE, err.Error())
					}
				}
				if err := db.Table("tx_order_actions").Delete(&orderActions).Error; err != nil {
					logx.WithContext(l.ctx).Error("搬移资料错误，删除tx_order_actions错误 : ", err.Error())
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			} else {
				break
			}

		}
		for i := 0; i <= 200; i++ {
			var orderFeeProfits []types.OrderFeeProfitX
			if err := db.Table("tx_orders_fee_profit").Where("created_at < ?", beforeDate).Limit(5000).Find(&orderFeeProfits).Error; err != nil {
				logx.WithContext(l.ctx).Error("搬移资料错误，查询tx_orders_fee_profit错误 : ", err.Error())
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			if len(orderFeeProfits) > 0 {
				for _, order := range orderFeeProfits {
					if err := db.Table("bu_orders_fee_profit").Create(order).Error; err != nil {
						logx.WithContext(l.ctx).Error("搬移资料错误，新增bu_orders_fee_profit错误 : ", err.Error())
						return errorz.New(response.DATABASE_FAILURE, err.Error())
					}
				}
				if err := db.Table("tx_orders_fee_profit").Delete(&orderFeeProfits).Error; err != nil {
					logx.WithContext(l.ctx).Error("搬移资料错误，查询tx_orders_fee_profit错误 : ", err.Error())
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			} else {
				break
			}

		}

		for i := 0; i <= 200; i++ {
			var orderChannels []types.OrderChannelsX
			if err := db.Table("tx_order_channels").Where("created_at < ?", beforeDate).Limit(5000).Find(&orderChannels).Error; err != nil {
				logx.WithContext(l.ctx).Error()
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			if len(orderChannels) > 0 {
				for _, order := range orderChannels {
					if err := db.Table("bu_order_channels").Create(order).Error; err != nil {
						logx.WithContext(l.ctx).Error()
						return errorz.New(response.DATABASE_FAILURE, err.Error())
					}
				}
				if err := db.Table("tx_order_channels").Delete(&orderChannels).Error; err != nil {
					logx.WithContext(l.ctx).Error()
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			} else {
				break
			}
		}
		return nil
	})

}
