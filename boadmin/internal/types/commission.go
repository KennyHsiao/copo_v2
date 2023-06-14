package types

import "com.copo/bo_service/common/gormx"

type CommissionMonthReportQueryAllResponseX struct {
	List                  []CommissionMonthReportX `json:"list"`
	TotalCommissionAmount float64                  `json:"totalCommission"`
	PageNum               int                      `json:"pageNum"`
	PageSize              int                      `json:"pageSize"`
	RowCount              int64                    `json:"rowCount"`
}

type CommissionMonthReportX struct {
	CommissionMonthReport
	ConfirmAt JsonTime `json:"confirmAt, optional"`
	CreatedAt JsonTime `json:"createdAt, optional"`
	UpdatedAt JsonTime `json:"createdAt, optional"`
}

type CommissionMonthReportDetailX struct {
	CommissionMonthReportDetail
	CreatedAt JsonTime `json:"createdAt, optional"`
	UpdatedAt JsonTime `json:"createdAt, optional"`
}

type CommissionMonthReportQueryAllRequestX struct {
	CommissionMonthReportQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type CommissionWithdrawOrderQueryAllRequestX struct {
	CommissionWithdrawOrderQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type CommissionWithdrawOrderX struct {
	CommissionWithdrawOrder
	CreatedAt JsonTime `json:"createdAt, optional"`
	UpdatedAt JsonTime `json:"updatedAt, optional"`
}

type CommissionWithdrawOrderQueryAllResponseX struct {
	List                []CommissionWithdrawOrderX `json:"list" gorm:"-"`
	TotalWithdrawAmount float64                    `json:"totalWithdrawAmount"`
	TotalPayAmount      float64                    `json:"totalPayAmount"`
	PageNum             int                        `json:"pageNum" gorm:"-"`
	PageSize            int                        `json:"pageSize" gorm:"-"`
	RowCount            int64                      `json:"rowCount" gorm:"-"`
}
