package channelpaytype

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelPayTypeQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeQueryLogic {
	return ChannelPayTypeQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeQueryLogic) ChannelPayTypeQuery(req types.ChannelPayTypeQueryRequest, language string) (resp *types.ChannelPayTypeQueryResponse, err error) {

	if err = l.svcCtx.MyDB.Table("ch_channel_pay_types ccpt").
		Select("ccpt.*, name_i18n->>'$."+language+"' as pay_type_name,cc.name as channel_name").
		Joins("LEFT JOIN ch_pay_types cpt ON ccpt.pay_type_code = cpt.code").
		Joins("LEFT JOIN ch_channels cc ON ccpt.channel_code = cc.code").
		Where("ccpt.id=?", req.ID).
		Take(&resp).Error; err != nil {
		logx.Error(err.Error())
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return
}
