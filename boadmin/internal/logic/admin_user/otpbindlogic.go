package admin_user

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/copo888/copo_otp/rpc/otpclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type OtpBindLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOtpBindLogic(ctx context.Context, svcCtx *svc.ServiceContext) OtpBindLogic {
	return OtpBindLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OtpBindLogic) OtpBind(req types.OtpBindRequest) (resp *types.OtpBindResponse, err error) {
	m := model.NewUser(l.svcCtx.MyDB)
	account, err := m.Exists(req.Account)
	if err != nil {
		return nil, err
	}

	if req.Password != "" {
		req.Password = utils.PasswordHash2(req.Password)

		l.svcCtx.MyDB.Table("au_users").
			Where("account = ?", req.Account).
			Updates(
				map[string]interface{}{
					"password": req.Password,
				},
			)

	}

	if req.Regenerate == "1" {

		res, err := l.svcCtx.OtpRpc.GenOtp(l.ctx, &otpclient.OtpGenRequest{
			Account: req.Account,
			Issuer:  "copo",
		})

		if err != nil {
			return nil, errorz.New(response.FAIL, err.Error())
		}

		l.svcCtx.MyDB.Table("au_users").Where("account = ?", req.Account).
			Updates(map[string]interface{}{"otp_key": res.Data.Secret, "qrcode": res.Data.Qrcode})

		return &types.OtpBindResponse{
			Secret: res.Data.Secret,
			Qrcode: fmt.Sprintf("%s%s", l.svcCtx.Config.ResourceHost, res.Data.Qrcode),
		}, nil
	}

	return &types.OtpBindResponse{
		Secret: account.OtpKey,
		Qrcode: fmt.Sprintf("%s%s", l.svcCtx.Config.ResourceHost, account.Qrcode),
	}, nil

}
