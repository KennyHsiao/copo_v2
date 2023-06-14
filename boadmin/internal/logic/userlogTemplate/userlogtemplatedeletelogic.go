package userlogTemplate

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogTemplateDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogTemplateDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogTemplateDeleteLogic {
	return UserLogTemplateDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogTemplateDeleteLogic) UserLogTemplateDelete(req *types.UserLogTemplateDeleteRequest) error {
	return l.svcCtx.MyDB.Table("au_user_log_template").Delete(&req).Error
}
