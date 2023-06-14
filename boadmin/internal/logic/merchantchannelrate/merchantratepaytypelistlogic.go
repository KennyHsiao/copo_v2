package merchantchannelrate

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantRatePayTypeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantRatePayTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantRatePayTypeListLogic {
	return MerchantRatePayTypeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantRatePayTypeListLogic) MerchantRatePayTypeList(req *types.MerchantRatePayTypeListRequest, language string) (resp *types.MerchantRatePayTypeListResponse, err error) {
	var payTypes []types.RatePayType
	//var terms []string
	db := l.svcCtx.MyDB.
		Table("mc_merchant_channel_rate as mcr").
		Joins("join ch_channels cc on mcr.channel_code = cc.code").
		Joins("join ch_pay_types pt on mcr.pay_type_code = pt.code")

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)

	if len(jwtMerchantCode) > 0 {
		db = db.Where("mcr.`merchant_code` = ?", jwtMerchantCode)
		//terms = append(terms, fmt.Sprintf(" mcr.`merchant_code` = '%s'", jwtMerchantCode))
	} else if len(req.MerchantCode) > 0 {
		db = db.Where("mcr.`merchant_code` = ?", req.MerchantCode)
		//terms = append(terms, fmt.Sprintf(" mcr.`merchant_code` = '%s'", req.MerchantCode))
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("cc.currency_code = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf("cc.currency_code = '%s'", req.CurrencyCode))
	}

	//term := strings.Join(terms, " AND ")

	if err = db.Select("distinct pt.code, name_i18n->>'$." + language + "' as name").
		Find(&payTypes).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, currency := range payTypes {
		if currency.Code == "DF" {
			payTypes = append([]types.RatePayType{currency}, append((payTypes)[:i], (payTypes)[i+1:]...)...)
			break
		}
	}

	resp = &types.MerchantRatePayTypeListResponse{
		List: payTypes,
	}

	return
}
