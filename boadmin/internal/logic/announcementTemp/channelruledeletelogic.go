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

type ChannelRuleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRuleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRuleDeleteLogic {
	return ChannelRuleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRuleDeleteLogic) ChannelRuleDelete(req *types.ChannelRuleDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		if err := db.Table("an_channel_rule").Delete(&req).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("删除渠道公告错误, %+v", err)
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}
