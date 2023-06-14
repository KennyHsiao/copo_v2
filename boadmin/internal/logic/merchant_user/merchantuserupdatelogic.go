package merchant_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUserUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUserUpdateLogic {
	return MerchantUserUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUserUpdateLogic) MerchantUserUpdate(req types.MerchantUserUpdateRequest) (err error) {
	var user types.User
	l.svcCtx.MyDB.Table("au_users").Where("id = ?", req.ID).Take(&user)

	update := types.User{
		ID:         req.ID,
		Account:    req.Account,
		Name:       req.Account,
		Phone:      req.Phone,
		Email:      req.Email,
		Roles:      req.Roles,
		Merchants:  req.Merchants,
		Currencies: req.Currencies,
		Status:     req.Status,
	}
	// 清除關聯
	_ = l.svcCtx.MyDB.Model(&user).Association("Roles").Clear()
	_ = l.svcCtx.MyDB.Model(&user).Association("Currencies").Clear()

	// Omit("Roles.*") 關閉自動維護upsert功能
	if err = l.svcCtx.MyDB.Table("au_users").
		Omit("Merchants.*").
		Omit("Roles.*").Updates(&types.UserX{User: update}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return err
}
