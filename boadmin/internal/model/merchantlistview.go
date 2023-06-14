package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
)

type MerchantListView struct {
	MyDB  *gorm.DB
	Table string
}

func NewMerchantListView(mydb *gorm.DB, t ...string) *MerchantListView {
	table := "merchant_list_view"
	if len(t) > 0 {
		table = t[0]
	}
	return &MerchantListView{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *MerchantListView) QueryListView(req types.MerchantQueryListViewRequestX) (resp *types.MerchantQueryListViewResponse, err error) {
	var merchants []types.MerchantListView
	var count int64
	//var terms []string
	db := m.MyDB
	db2 := m.MyDB
	if len(req.JwtMerchantCode) > 0 { // 代理報表

		var SubAgentsMerchants []types.Merchant
		var merchantCodes []string
		if SubAgentsMerchants, err = NewMerchant(m.MyDB).GetSubAgents(req.JwtMerchantCode); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		for _, m := range SubAgentsMerchants {
			merchantCodes = append(merchantCodes, m.Code)
		}
		//terms = append(terms, fmt.Sprintf(" code in ('%s')", strings.Join(merchantCodes, "','")))
		db = db.Where("code in ?", merchantCodes)
		db2 = db2.Where("code in ?", merchantCodes)
	}
	if len(req.AccountName) > 0 {
		//terms = append(terms, fmt.Sprintf("`account_name` like '%%%s%%'", req.AccountName))
		db = db.Where("account_name like ?", "%"+req.AccountName+"%")
		db2 = db2.Where("account_name like ?", "%"+req.AccountName+"%")
	}
	if len(req.Code) > 0 {
		//terms = append(terms, fmt.Sprintf("code like '%%%s%%'", req.Code))
		db = db.Where("code like ?", "%"+req.Code+"%")
		db2 = db2.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.AgentLayerCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`agent_layer_code` like '%%%s%%'", req.AgentLayerCode))
		db = db.Where("agent_layer_code like ?", "%"+req.AgentLayerCode+"%")
		db2 = db2.Where("agent_layer_code like ?", "%"+req.AgentLayerCode+"%")
	}
	if len(req.BalanceCurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`currency_code` = '%s'", req.BalanceCurrencyCode))
		db = db.Where("currency_code = ?", req.BalanceCurrencyCode)
		db2 = db2.Where("currency_code = ?", req.BalanceCurrencyCode)
	}
	if len(req.Currencies) > 0 {
		//terms = append(terms, fmt.Sprintf("`currency_code` in ('%s') ", strings.Join(req.Currencies, "','")))
		db = db.Where("currency_code in ?", req.Currencies)
		db2 = db2.Where("currency_code in ?", req.Currencies)
	}
	if len(req.MerchantCurrenciesStatus) > 0 {
		//terms = append(terms, fmt.Sprintf("`merchant_currencies_status` = '%s'", req.MerchantCurrenciesStatus))
		db = db.Where("merchant_currencies_status = ?", req.MerchantCurrenciesStatus)
		db2 = db2.Where("merchant_currencies_status = ?", req.MerchantCurrenciesStatus)
	}

	//term := strings.Join(terms, " AND ")
	if err = db2.Table(m.Table).Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table(m.Table).
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&merchants).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantQueryListViewResponse{
		List:     merchants,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}

func (m *MerchantListView) QueryListViewTotal(req types.MerchantQueryListViewRequestX) (resp *types.MerchantQueryListViewTotalResponse, err error) {
	//var terms []string
	db := m.MyDB
	if len(req.JwtMerchantCode) > 0 { // 代理報表

		var SubAgentsMerchants []types.Merchant
		var merchantCodes []string
		if SubAgentsMerchants, err = NewMerchant(m.MyDB).GetSubAgents(req.JwtMerchantCode); err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		for _, m := range SubAgentsMerchants {
			merchantCodes = append(merchantCodes, m.Code)
		}
		//terms = append(terms, fmt.Sprintf(" code in ('%s')", strings.Join(merchantCodes, "','")))
		db = db.Where("code in ?", merchantCodes)
	}
	if len(req.AccountName) > 0 {
		//terms = append(terms, fmt.Sprintf("`account_name` like '%%%s%%'", req.AccountName))
		db = db.Where("account_name like ?", "%"+req.AccountName+"%")
	}
	if len(req.Code) > 0 {
		//terms = append(terms, fmt.Sprintf("code like '%%%s%%'", req.Code))
		db = db.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.AgentLayerCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`agent_layer_code` like '%%%s%%'", req.AgentLayerCode))
		db = db.Where("agent_layer_code like ?", "%"+req.AgentLayerCode+"%")
	}
	if len(req.BalanceCurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`currency_code` = '%s'", req.BalanceCurrencyCode))
		db = db.Where("currency_code = ?", req.BalanceCurrencyCode)
	}
	if len(req.Currencies) > 0 {
		//terms = append(terms, fmt.Sprintf("`currency_code` in ('%s') ", strings.Join(req.Currencies, "','")))
		db = db.Where("currency_code in ?", req.Currencies)
	}
	if len(req.MerchantCurrenciesStatus) > 0 {
		//terms = append(terms, fmt.Sprintf("`merchant_currencies_status` = '%s'", req.MerchantCurrenciesStatus))
		db = db.Where("merchant_currencies_status = ?", req.MerchantCurrenciesStatus)
	}

	//term := strings.Join(terms, " AND ")

	seletStr := "sum(xf_balance) as xf_balance_total," +
		"sum(df_balance) as df_balance_total," +
		"sum(total_balance) as balance_total," +
		"sum(commission) as commission_total," +
		"sum(frozen_amount) as frozen_amount_total"
	if err = db.Select(seletStr).Table(m.Table).
		Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
