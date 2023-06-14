package login

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/copo_otp/rpc/otpclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type OtpValidLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOtpValidLogic(ctx context.Context, svcCtx *svc.ServiceContext) OtpValidLogic {
	return OtpValidLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OtpValidLogic) OtpValid(req types.OtpValidRequest) (resp *types.OtpValidResponse, err error) {
	user := struct {
		ID     int64
		OtpKey string
	}{}

	account := l.ctx.Value("account").(string)

	if err = l.svcCtx.MyDB.Table("au_users").Where("account = ?", account).Take(&user).Error; err != nil {
		return nil, errorz.New(response.USER_NOT_FOUND, err.Error())
	}

	res, err := l.svcCtx.OtpRpc.Validate(l.ctx, &otpclient.OtpVaildRequest{
		PassCode: req.Otp,
		Secret:   user.OtpKey,
	})

	if err != nil {
		return nil, errorz.New(response.SERVICE_PROVIDER_NOT_FOUND, err.Error())
	}

	return &types.OtpValidResponse{
		Valid: res.Vaild,
	}, nil
}
