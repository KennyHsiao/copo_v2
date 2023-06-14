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

type MerchantQueryRateListViewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryRateListViewLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryRateListViewLogic {
	return MerchantQueryRateListViewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryRateListViewLogic) MerchantQueryRateListView(req types.MerchantQueryRateListViewRequestX) (resp *types.MerchantQueryRateListViewResponse, err error) {
	var merchantRates []types.MerchantRateListView
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("merchant_rate_list_view")

	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("currency_code = '%s'", req.CurrencyCode))
		db = db.Where("currency_code = ?", req.CurrencyCode)
	}
	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("mer_code like '%%%s%%'", req.MerchantCode))
		db = db.Where("mer_code like ?", "%"+req.MerchantCode+"%")
	}
	if len(req.ParentMerchantCode) > 0 {
		db = db.Where("parent_mer_code like ?", "%"+req.ParentMerchantCode+"%")
		//terms = append(terms, fmt.Sprintf("parent_mer_code like '%%%s%%'", req.ParentMerchantCode))
	}
	if len(req.PayTypeCode) > 0 {
		db = db.Where("pay_type_code = ?", req.PayTypeCode)
		//terms = append(terms, fmt.Sprintf("pay_type_code = '%s'", req.PayTypeCode))
	}
	if len(req.PayTypeName) > 0 {
		db = db.Where("pay_type_name like ?", "%"+req.PayTypeName+"%")
		//terms = append(terms, fmt.Sprintf("pay_type_name like '%%%s%%'", req.PayTypeName))
	}
	if len(req.ChannelName) > 0 {
		db = db.Where("chn_name like ?", "%"+req.ChannelName+"%")
		//terms = append(terms, fmt.Sprintf("chn_name like '%%%s%%'", req.ChannelName))
	}
	if len(req.ChannelPayTypeCode) > 0 {
		db = db.Where("chn_pay_type_code like ?", "%"+req.ChannelPayTypeCode+"%")
		//terms = append(terms, fmt.Sprintf("chn_pay_type_code like '%%%s%%'", req.ChannelPayTypeCode))
	}
	db = db.Where("chn_status = '1'")
	db = db.Where("chn_pay_type_status = '1'")
	db = db.Where("mer_status = '1'")
	db = db.Where("mer_chn_rate_status = '1'")
	//terms = append(terms, "chn_status = '1'")
	//terms = append(terms, "chn_pay_type_status = '1'")
	//terms = append(terms, "mer_status = '1'")
	//terms = append(terms, "mer_chn_rate_status = '1'")

	//term := strings.Join(terms, " AND ")
	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&merchantRates).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantQueryRateListViewResponse{
		List:     merchantRates,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
