package merchantchannelrate

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantGetParentCodeDropDownListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantGetParentCodeDropDownListLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantGetParentCodeDropDownListLogic {
	return MerchantGetParentCodeDropDownListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantGetParentCodeDropDownListLogic) MerchantGetParentCodeDropDownList(req types.MerchantRateGetParentCodeDropDownListRequest) (resp []string, err error) {
	//var terms []string

	db := l.svcCtx.MyDB.Table("merchant_rate_list_view")

	if req.CurrencyCode != "" {
		db = db.Where("currency_code =  ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf("currency_code = '%s'", req.CurrencyCode))
	}

	db = db.Where("chn_status = '1'")
	db = db.Where("chn_pay_type_status = '1'")
	db = db.Where("mer_status = '1'")
	db = db.Where("mer_chn_rate_status = '1'")
	db = db.Where("parent_mer_code != ''")
	db = db.Where("parent_mer_code is not null")

	//terms = append(terms, "chn_status = '1'")
	//terms = append(terms, "chn_pay_type_status = '1'")
	//terms = append(terms, "mer_status = '1'")
	//terms = append(terms, "mer_chn_rate_status = '1'")
	//terms = append(terms, "parent_mer_code != '' ")
	//terms = append(terms, "parent_mer_code is not null")
	//term := strings.Join(terms, " AND ")

	if err = db.
		Order("parent_mer_code asc").
		Distinct().
		Pluck("parent_mer_code", &resp).Error; err != nil {

		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
