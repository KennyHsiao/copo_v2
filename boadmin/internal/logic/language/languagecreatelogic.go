package language

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LanguageCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLanguageCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) LanguageCreateLogic {
	return LanguageCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LanguageCreateLogic) LanguageCreate(req types.LanguageCreateRequest) error {

	language := &types.LanguageCreate{
		LanguageCreateRequest: req,
	}
	return l.svcCtx.MyDB.Table("bs_lang").Create(language).Error
}
