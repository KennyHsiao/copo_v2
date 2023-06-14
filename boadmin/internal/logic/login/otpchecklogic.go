package login

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/copo_otp/rpc/otpclient"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OtpCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOtpCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) OtpCheckLogic {
	return OtpCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OtpCheckLogic) OtpCheck(req types.OtpCheckRequest) (err error) {

	user := struct {
		ID      int64
		OtpKey  string
		IsLogin string
	}{}

	account := req.Account

	if err = l.svcCtx.MyDB.Table("au_users").Where("account = ?", account).Take(&user).Error; err != nil {
		return errorz.New(response.USER_NOT_FOUND, err.Error())
	}

	res, err := l.svcCtx.OtpRpc.Validate(l.ctx, &otpclient.OtpVaildRequest{
		PassCode: req.Otp,
		Secret:   user.OtpKey,
	})

	if err != nil {
		return errorz.New(response.SERVICE_PROVIDER_NOT_FOUND, err.Error())
	}

	if res.Vaild {
		if err := l.svcCtx.MyDB.Table("au_users").Where("account = ?", account).
			Updates(map[string]interface{}{
				"is_bind": "1",
			}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	}

	return errorz.New(response.VERIFICATION_CODE_EXPIRED)
}
