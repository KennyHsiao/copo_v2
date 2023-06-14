package admin_user

import (
	"com.copo/bo_service/boadmin/internal/model"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserMenuLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserMenuLogic {
	return UserMenuLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserMenuLogic) UserMenu(req types.UserMenuRequest) (resp *types.UserMenuResponseX, err error) {
	m := model.NewUser(l.svcCtx.MyDB)
	v, err := m.Menu(l.ctx.Value("account").(string))

	if err != nil {
		return nil, err
	}

	// 過濾permits
	userPermits := []model.UserPermit{}
	l.svcCtx.MyDB.Table("au_role_permits").Where("role_id", v.Roles[0].ID).Find(&userPermits)

	userPermitMap := map[int64]bool{}

	for _, p := range userPermits {
		userPermitMap[p.PermitId] = true
	}

	return &types.UserMenuResponseX{MenuTree: types.GenMenuTreeFilter(v.Roles[0].Menus, userPermitMap)}, nil
}
