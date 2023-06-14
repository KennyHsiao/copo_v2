package merchant

import (
	"com.copo/bo_service/common/utils"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantResetPasswordLogic {
	return MerchantResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantResetPasswordLogic) MerchantResetPassword(req types.MerchantResetPasswordRequest) error {

	err := l.svcCtx.MyDB.Table("au_users").Where("account = ?", req.Name).
		Update("password", utils.PasswordHash2("00000000")).Error

	return err
}
