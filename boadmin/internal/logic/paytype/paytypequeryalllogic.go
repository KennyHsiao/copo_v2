package paytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeQueryAllLogic {
	return PayTypeQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeQueryAllLogic) PayTypeQueryAll(req types.PayTypeQueryAllRequestX) (resp *types.PayTypeQueryAllResponse, err error) {
	return model.NewPayType(l.svcCtx.MyDB).PayTypeQueryAll(req)
}
