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

type ChannelRuleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRuleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRuleCreateLogic {
	return ChannelRuleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRuleCreateLogic) ChannelRuleCreate(req types.ChannelRuleCreateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		channelRule := &types.ChannelRuleCreate{
			ChannelRuleCreateRequest: req,
		}

		if err := db.Table("an_channel_rule").Create(channelRule).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})

}
