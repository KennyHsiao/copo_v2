package timezonex

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimezonexQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimezonexQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) TimezonexQueryLogic {
	return TimezonexQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimezonexQueryLogic) TimezonexQuery(req types.TimezonexQueryRequest) (resp *types.TimezonexQueryResponse, err error) {
	err = l.svcCtx.MyDB.Table("bs_timezone").First(&resp, req.ID).Error
	return
}
