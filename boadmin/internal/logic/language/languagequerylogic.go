package language

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LanguageQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLanguageQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) LanguageQueryLogic {
	return LanguageQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LanguageQueryLogic) LanguageQuery(req types.LanguageQueryRequest) (resp *types.LanguageQueryResponse, err error) {

	err = l.svcCtx.MyDB.Table("bs_lang").First(&resp, req.ID).Error
	return

}
