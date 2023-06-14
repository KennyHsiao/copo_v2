package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ChannelPayTypeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeDeleteLogic {
	return ChannelPayTypeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeDeleteLogic) ChannelPayTypeDelete(req types.ChannelPayTypeDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		return tx.Table("ch_channel_pay_types").Delete(&req).Error
	})
}
