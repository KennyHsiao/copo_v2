package admin_user

import (
	"com.copo/bo_service/common/utils"
	"context"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserRefreshTokenLogic {
	return UserRefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRefreshTokenLogic) UserRefreshToken(req types.UserRefreshTokenRequest) (resp *types.UserRefreshTokenResponse, err error) {

	accessExpire := l.svcCtx.Config.Auth.AccessExpire
	now := time.Now().Unix()

	payloads := make(map[string]interface{})
	payloads["userId"] = l.ctx.Value("userId")
	payloads["account"] = l.ctx.Value("account")
	payloads["merchantCode"] = l.ctx.Value("merchantCode")
	payloads["isAdmin"] = l.ctx.Value("isAdmin")
	payloads["name"] = l.ctx.Value("name")
	payloads["identity"] = l.ctx.Value("identity")
	accessToken, err := utils.GenToken(now, l.svcCtx.Config.Auth.AccessSecret, payloads, accessExpire)
	if err != nil {
		return nil, err
	}

	return &types.UserRefreshTokenResponse{
		AccessToken:  accessToken,
		AccessExpire: now + accessExpire,
		RefreshAfter: now + accessExpire/2,
	}, nil
}
