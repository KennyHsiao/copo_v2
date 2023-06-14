package admin_user

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserUnfreezeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUnfreezeLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserUnfreezeLogic {
	return UserUnfreezeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUnfreezeLogic) UserUnfreeze(req *types.UserUnfreezeRequest) (err error) {

	var user types.User

	if err = l.svcCtx.MyDB.Table("au_users").Where("account = ?", req.Account).Take(&user).Error; err != nil {
		return errorz.New(response.USER_NOT_FOUND)
	}

	if err = model.NewUser(l.svcCtx.MyDB).UpdateFailCount(user.Account, 0); err != nil {
		return errorz.New(response.USER_NOT_FOUND)
	}

	return
}
