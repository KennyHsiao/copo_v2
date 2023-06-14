package types

import "com.copo/bo_service/common/gormx"

type PayCheckBillQueryRequestX struct {
	PayCheckBillQueryRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type WithdrawCheckBillQueryRequestX struct {
	WithdrawCheckBillQueryRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type MerchantReportQueryRequestX struct {
	MerchantReportQueryRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type ChannelReportQueryRequestX struct {
	ChannelReportQueryRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type TotalMerchantBalances struct {
	DfBalances float64 `json:"df_balances"`
	XfBalances float64 `json:"xf_balances"`
}

type WeeklyTransDetailResponseX struct {
	List []*WeeklyTransDetail `json:"list"`
}
