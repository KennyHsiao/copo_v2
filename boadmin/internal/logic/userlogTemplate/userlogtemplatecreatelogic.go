package userlogTemplate

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogTemplateCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogTemplateCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogTemplateCreateLogic {
	return UserLogTemplateCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogTemplateCreateLogic) UserLogTemplateCreate(req types.UserLogTemplateCreateRequest) error {
	temp := &types.UserLogTemplateCreate{
		UserLogTemplateCreateRequest: req,
	}

	return l.svcCtx.MyDB.Table("au_user_log_template").Create(temp).Error
}
