package merchantchannelrate

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantConfigureRateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantConfigureRateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantConfigureRateListLogic {
	return MerchantConfigureRateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantConfigureRateListLogic) MerchantConfigureRateList(req types.MerchantConfigureRateListRequestX) (resp *types.MerchantConfigureRateListResponse, err error) {
	var merchantConfigureRates []types.MerchantConfigureRate
	//var terms []string

	db := l.svcCtx.MyDB.
		Table("ch_channel_pay_types as ccpt").
		Joins("join ch_channels cc on ccpt.channel_code = cc.code").
		Joins("join bs_currencies bc on cc.currency_code = bc.code").
		Joins("join ch_pay_types cpt on ccpt.pay_type_code = cpt.code").
		Joins("join mc_merchants mm on mm.code = ? ", req.MerchantCode).
		Joins("left join mc_merchant_channel_rate mmcr on ccpt.code = mmcr.channel_pay_types_code and mmcr.merchant_code = mm.code").
		Joins("left join mc_merchant_channel_rate mmcrp on ccpt.code = mmcrp.channel_pay_types_code and mm.agent_parent_code = mmcrp.merchant_code").
		Joins("left join mc_merchant_pt_balances mmpb on mmpb.id = mmcr.merchant_pt_balance_id")
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)

	selectX := " IF(mmcr.merchant_code IS NULL, 'false', 'true') as is_configure," +
		"cc.name                   as chn_name," +
		"cc.code                   as chn_code," +
		"bc.code                   as currency_code," +
		"bc.name                   as currency_name," +
		"cc.is_proxy               as chn_is_proxy," +
		"cc.is_nz_pre              as chn_is_nz_pre," +
		"cc.status                 as chn_status," +
		"cpt.code                  as pay_type_code," +
		"cpt.name_i18n->>'$." + req.Language + "' as pay_type_name," +
		"ccpt.code                 as chn_pay_type_code," +
		"IF(mm.agent_parent_code = '', ccpt.fee, mmcrp.fee)                    as pay_type_fee," +
		"IF(mm.agent_parent_code = '', ccpt.handling_fee, mmcrp.handling_fee)  as pay_type_handling_fee," +
		"ccpt.max_internal_charge  as chn_pay_type_max_internal_charge," +
		"ccpt.daily_tx_limit       as chn_pay_type_daily_tx_limit," +
		"ccpt.single_min_charge    as chn_pay_type_single_min_charge," +
		"ccpt.single_max_charge    as chn_pay_type_single_max_charge," +
		"ccpt.bill_date            as chn_pay_type_bill_date," +
		"ccpt.status               as chn_pay_type_status," +
		"mm.code                   as mer_code," +
		"mm.bill_lading_type       as mer_bill_lading_type," +
		"mm.status                 as mer_status," +
		"mmcr.id                   as mer_chn_rate_id," +
		"mmcr.fee                  as mer_chn_rate_fee," +
		"mmcr.handling_fee         as mer_chn_rate_handling_fee," +
		"mmcr.designation          as mer_chn_rate_designation," +
		"mmcr.designation_no       as mer_chn_rate_designation_no," +
		"mmcr.status               as mer_chn_rate_status," +
		"mmcr.merchant_pt_balance_id   as merchant_pt_balance_id," +
		"mmpb.name                 as merchant_pt_balance_name," +
		"mmcrp.merchant_code       as parent_mer_code," +
		"mmcrp.status              as parent_mer_chn_rate_status"

	if len(jwtMerchantCode) > 0 {
		// 商戶帳號只能看見啟用渠道

		db = db.Where("ccpt.status = '1'")
		db = db.Where("cc.status = '1'")

		//terms = append(terms, "ccpt.status = '1' ")
		//terms = append(terms, "cc.status = '1' ")
		// 商戶帳號只能看見已配置
		//terms = append(terms, "mmcr.id  is not null")
		db = db.Where("mmcr.id  is not null")

	} else {

		// 系統方只排除關閉的
		//terms = append(terms, "ccpt.status != '0' ")
		//terms = append(terms, "cc.status != '0' ")
		db = db.Where("ccpt.status != '0'")
		db = db.Where("cc.status != '0'")

	}

	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("cc.currency_code = '%s'", req.CurrencyCode))
		db = db.Where("cc.currency_code = ?", req.CurrencyCode)
	}

	if len(req.ChannelCode) > 0 {
		//terms = append(terms, fmt.Sprintf("cc.code = '%s'", req.ChannelCode))
		db = db.Where("cc.code = = ?", req.ChannelCode)
	}

	if req.PayTypeCode == "00" { //  指定渠道
		//terms = append(terms, "mmcr.designation = '1'")
		db = db.Where("mmcr.designation = '1'")

	} else if req.PayTypeCode == "01" { // 配置渠道
		//terms = append(terms, "mmcr.id  is not null")
		db = db.Where("mmcr.id  is not null")
		// 配置渠道只能看見啟用渠道
		//db = db.Where("ccpt.status = '1'")
		//db = db.Where("cc.status = '1'")
		//terms = append(terms, "ccpt.status = '1' ")
		//terms = append(terms, "cc.status = '1' ")
	} else {
		db = db.Where("cpt.code = ?", req.PayTypeCode)
		//terms = append(terms, fmt.Sprintf("cpt.code = '%s'", req.PayTypeCode))
	}

	//term := strings.Join(terms, " AND ")

	if err = db.Select(selectX).
		Scopes(gormx.Sort(req.Orders)).
		Find(&merchantConfigureRates).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantConfigureRateListResponse{
		List: merchantConfigureRates,
	}

	return
}
