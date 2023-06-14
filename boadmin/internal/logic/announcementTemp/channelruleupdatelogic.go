package announcementTemp

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelRuleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRuleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRuleUpdateLogic {
	return ChannelRuleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRuleUpdateLogic) ChannelRuleUpdate(req types.ChannelRuleUpdateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		channelRuleUpdate := types.ChannelRuleUpdate{
			ChannelRuleUpdateRequest: req,
		}

		if err := db.Table("an_channel_rule").Where("id = ?", channelRuleUpdate.ID).Updates(channelRuleUpdate).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})

}
