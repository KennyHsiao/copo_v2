package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelPayTypeMerchantRateQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeMerchantRateQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeMerchantRateQueryLogic {
	return ChannelPayTypeMerchantRateQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeMerchantRateQueryLogic) ChannelPayTypeMerchantRateQuery(req *types.MerchantRateQueryRequestX) (resp *types.MerchantRateQueryResponse, err error) {
	merChannelRate := &[]types.MerConfigureRate{}
	var count int64
	db := l.svcCtx.MyDB

	if len(req.ChannelPayTypeCode) > 0 {
		db = db.Where("mmcr.channel_pay_types_code = ? ", req.ChannelPayTypeCode)
	}

	selectX :=
		"mmcr.channel_code         as channel_code," +
			"mmcr.pay_type_code        as pay_type_code," +
			"bc.code                   as currency_code," +
			"ccpt.code                 as chn_pay_type_code," +
			"ccpt.fee                  as pay_type_fee," +
			"ccpt.handling_fee         as pay_type_handling_fee," +
			"ccpt.single_min_charge    as chn_pay_type_single_min_charge," +
			"mm.code                   as mer_code," +
			"mm.rate_check             as rate_check," +
			"mm.bill_lading_type       as mer_bill_lading_type," +
			"mmcr.id                   as mer_chn_rate_id," +
			"mmcr.fee                  as mer_chn_rate_fee," +
			"mmcr.handling_fee         as mer_chn_rate_handling_fee," +
			"mmcr.designation          as mer_chn_rate_designation," +
			"mmcr.designation_no       as mer_chn_rate_designation_no"

	tx := db.Table("mc_merchant_channel_rate AS mmcr ").
		Joins("LEFT JOIN mc_merchants mm ON mm.CODE = mmcr.merchant_code ").
		Joins("LEFT JOIN ch_channel_pay_types ccpt ON ccpt.CODE = mmcr.channel_pay_types_code ").
		Joins("JOIN ch_channels cc ON ccpt.channel_code = cc.CODE ").
		Joins("JOIN bs_currencies bc ON cc.currency_code = bc.CODE ").
		Joins("JOIN ch_pay_types cpt ON ccpt.pay_type_code = cpt.CODE ").
		Scopes(gormx.Sort(req.Orders))

	if err = tx.Select(selectX).Find(&merChannelRate).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = tx.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return &types.MerchantRateQueryResponse{
		List:     *merChannelRate,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}, nil

	return
}
