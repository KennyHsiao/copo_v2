package userlogTemplate

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogTemplateQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogTemplateQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogTemplateQueryLogic {
	return UserLogTemplateQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogTemplateQueryLogic) UserLogTemplateQuery(req *types.UserLogTemplateQueryRequest) (resp *types.UserLogTemplateQueryResponseX, err error) {
	err = l.svcCtx.MyDB.Table("au_user_log_template").Take(&resp, req.ID).Error
	return
}
