package paytype

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeImageUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeImageUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeImageUploadLogic {
	return PayTypeImageUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}
