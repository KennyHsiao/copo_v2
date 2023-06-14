package admin_user

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserDeleteLogic {
	return UserDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserDeleteLogic) UserDelete(req types.UserDeleteRequest) error {
	resp, err := model.NewUser(l.svcCtx.MyDB).GetUser(req.ID)
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	if resp.DisableDelete == "1" {
		return errorz.New(response.ACCOUNT_DISABLE_DELETE, "此帳號禁止刪除")
	}

	return l.svcCtx.MyDB.Table("au_users").Delete(&req).Error
}
