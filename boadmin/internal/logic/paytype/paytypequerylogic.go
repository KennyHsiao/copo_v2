package paytype

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeQueryLogic {
	return PayTypeQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeQueryLogic) PayTypeQuery(req types.PayTypeQueryRequest) (resp *types.PayTypeQueryResponse, err error) {
	if err = l.svcCtx.MyDB.Table("ch_pay_types").Take(&resp, req.ID).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	if len(resp.ImgUrl) > 0 {
		resp.PayType.ImgUrl = l.svcCtx.Config.ResourceHost + resp.ImgUrl
	} else {
		resp.PayType.ImgUrl = ""
	}

	return
}
