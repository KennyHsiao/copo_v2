package channelraterecord

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelRateRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRateRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRateRecordQueryAllLogic {
	return ChannelRateRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRateRecordQueryAllLogic) ChannelRateRecordQueryAll(req *types.ChannelRateRecordRequestX) (resp *types.ChannelRateRecordresponse, err error) {
	var recordList []types.ChannelRateRecord
	var count int64

	l.svcCtx.MyDB.Table("ch_channel_rate_record").Where("channel_pay_type_code LIKE ?", "%"+req.ChannelCode+"%").Count(&count)

	if err = l.svcCtx.MyDB.Table("ch_channel_rate_record").
		Scopes(gormx.Paginate(req.ChannelRateRecordRequest)).
		Scopes(gormx.Sort(req.Orders)).
		Where("channel_pay_type_code LIKE ?", "%"+req.ChannelCode+"%").
		Find(&recordList).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, record := range recordList {
		recordList[i].CreatedAt = utils.ParseTime(record.CreatedAt)
	}

	return &types.ChannelRateRecordresponse{
		List:     recordList,
		PageSize: req.PageSize,
		PageNum:  req.PageNum,
		RowCount: count,
	}, nil
}
