package language

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LanguageUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLanguageUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) LanguageUpdateLogic {
	return LanguageUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LanguageUpdateLogic) LanguageUpdate(req types.LanguageUpdateRequest) error {

	lang := &types.LanguageUpdate{
		LanguageUpdateRequest: req,
	}
	return l.svcCtx.MyDB.Table("bs_lang").Updates(lang).Error
}
