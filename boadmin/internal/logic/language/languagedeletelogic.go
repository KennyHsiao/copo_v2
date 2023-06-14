package language

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LanguageDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLanguageDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) LanguageDeleteLogic {
	return LanguageDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LanguageDeleteLogic) LanguageDelete(req types.LanguageDeleteRequest) error {
	return l.svcCtx.MyDB.Table("bs_lang").Delete(&req).Error
}
