package merchantraterecord

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantRateRecordQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantRateRecordQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantRateRecordQueryAllLogic {
	return MerchantRateRecordQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantRateRecordQueryAllLogic) MerchantRateRecordQueryAll(req types.MerchantRateRecordRequestX) (resp *types.MerchantRateRecordResponse, err error) {
	var recordList []types.MerchantRateRecord
	var count int64
	if err = l.svcCtx.MyDB.Table("mc_merchant_rate_record").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Where("channel_pay_type_code = ?", req.ChannelPayTypeCode).
		Find(&recordList).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	l.svcCtx.MyDB.Table("mc_merchant_rate_record").Where("channel_pay_type_code = ?", req.ChannelPayTypeCode).Count(&count)

	for i, record := range recordList {
		recordList[i].CreatedAt = utils.ParseTime(record.CreatedAt)
	}

	return &types.MerchantRateRecordResponse{
		List:     recordList,
		PageSize: req.PageSize,
		PageNum:  req.PageNum,
		RowCount: count,
	}, nil
}
