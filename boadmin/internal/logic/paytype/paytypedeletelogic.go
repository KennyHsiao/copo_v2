package paytype

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeDeleteLogic {
	return PayTypeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeDeleteLogic) PayTypeDelete(req types.PayTypeDeleteRequest) error {
	if err := l.svcCtx.MyDB.Table("ch_pay_types").Delete(&req).Error; err != nil {
		return errorz.New(response.DELETE_DATABASE_FAILURE, err.Error())
	}
	return nil
}
