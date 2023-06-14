package permit

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"gorm.io/gorm/clause"

	"github.com/zeromicro/go-zero/core/logx"
)

type PermitUpsertLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPermitUpsertLogic(ctx context.Context, svcCtx *svc.ServiceContext) PermitUpsertLogic {
	return PermitUpsertLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PermitUpsertLogic) PermitUpsert(req types.PermitUpsertRequest) error {
	menuId := req.MenuID

	for _, permit := range req.Permits {
		permit.MenuID = menuId
		l.svcCtx.MyDB.Table("au_permits").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "need_password"}),
		}).Create(&permit)
	}

	return nil
}
