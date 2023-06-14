package userlogTemplate

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogTemplateUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogTemplateUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogTemplateUpdateLogic {
	return UserLogTemplateUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogTemplateUpdateLogic) UserLogTemplateUpdate(req types.UserLogTemplateUpdateRequest) error {
	temp := &types.UserLogTemplateUpdate{
		UserLogTemplateUpdateRequest: req,
	}
	return l.svcCtx.MyDB.Table("au_user_log_template").Updates(temp).Error
}
