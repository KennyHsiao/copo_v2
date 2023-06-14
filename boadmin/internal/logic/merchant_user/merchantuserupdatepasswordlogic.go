package merchant_user

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUserUpdatePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUserUpdatePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUserUpdatePasswordLogic {
	return MerchantUserUpdatePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUserUpdatePasswordLogic) MerchantUserUpdatePassword(req types.MerchantUserUpdatePasswordRequest) (err error) {

	var user types.User

	if err = l.svcCtx.MyDB.Table("au_users").Where("account = ?", req.Account).Take(&user).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if !utils.CheckPassword2(req.OldPassword, user.Password) {
		return errorz.New(response.INCORRECT_USER_PASSWD, "密碼錯誤")
	}

	//user.Password = utils.PasswordHash(req.Password)
	//user.IsLogin = "1"

	if err = l.svcCtx.MyDB.Table("au_users").
		Where("account = ?", user.Account).
		Updates(map[string]interface{}{
			"password": utils.PasswordHash2(req.Password),
			"is_login": "1",
		}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return nil
}
