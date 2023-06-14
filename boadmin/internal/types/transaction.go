package types

import (
	"com.copo/bo_service/common/gormx"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
	"unsafe"
)

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	var res string
	if !time.Time(j).IsZero() {
		var stamp = fmt.Sprintf("%s", time.Time(j).Format("2006-01-02 15:04:05"))
		str := strings.Split(stamp, " +")
		res = str[0]
		return json.Marshal(res)
	}
	return json.Marshal("")
}

func (j JsonTime) Time() time.Time {
	return time.Time(j)
}

func (j JsonTime) Value() (driver.Value, error) {
	return time.Time(j), nil
}

func (j JsonTime) Parse(s string, zone ...string) (JsonTime, error) {

	var (
		loc *time.Location
		err error
	)
	if len(zone) > 0 {
		loc, err = time.LoadLocation(zone[0])
	} else {
		loc, err = time.LoadLocation("")
	}

	if err != nil {
		return j, err
	}

	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err != nil {
		return j, err
	}
	jt := (*JsonTime)(unsafe.Pointer(&t))

	return *jt, nil
}

func (j JsonTime) New(ts ...time.Time) JsonTime {
	var t time.Time

	if len(ts) > 0 {
		t = ts[0]
	} else {
		t = time.Now().UTC()
	}

	jt := (*JsonTime)(unsafe.Pointer(&t))
	return *jt
}

func (j JsonTime) FormatAndZero(layout string, zoneName string) string {
	if !time.Time(j).IsZero() {
		zone, _ := time.LoadLocation(zoneName)
		return j.Time().UTC().In(zone).Format(layout)
	}
	return ""
}

func (OrderChannels) TableName() string {
	return "tx_order_channels"
}

type OrderX struct {
	Order
	TransAt           JsonTime `json:"transAt, optional"`
	FrozenAt          JsonTime `json:"frozenAt, optiona"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ChannelCallBackAt time.Time `json:"channelCallBackAt"`
}

type OrderD struct {
	Order
	ProfitAmount float64 `json:"profitAmount"`
}

type OrderInternalCreate struct {
	OrderX
	FormData map[string][]*multipart.FileHeader `gorm:"-"`
}

type OrderInternalUpdate struct {
	OrderX
}

type OrderWithdrawUpdate struct {
	OrderX
}

type UploadImageRequestX struct {
	UploadImageRequest
	UploadFile   multipart.File
	UploadHeader *multipart.FileHeader
	Files        map[string][]*multipart.FileHeader
}

type MerchantRateListViewX struct {
	MerchantRateListView
	Balance float64 `json:"balance"`
}

type MerchantOrderRateListViewX struct {
	MerchantOrderRateListView
	Balance float64 `json:"balance"`
}

type OrderQueryMerchantCurrencyAndBanks struct {
	MerchantOrderRateListViewX *MerchantOrderRateListViewX `json:"merchantOrderRateListViewX"`
	ChannelBanks               []ChannelBankX
}

type OrderQueryMerchantCurrencyAndBanksResponseX struct {
	List []OrderQueryMerchantCurrencyAndBanks `json:"list"`
}

type ReceiptRecordQueryAllRequestX struct {
	ReceiptRecordQueryAllRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type IncomeExpenseQueryRequestX struct {
	IncomeExpenseQueryRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type PtBalanceRecordsQueryRequestX struct {
	PtBalanceRecordsQueryRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type FrozenRecordQueryAllRequestX struct {
	FrozenRecordQueryAllRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type DeductRecordQueryAllRequestX struct {
	DeductRecordQueryAllRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type AllocRecordQueryAllRequestX struct {
	AllocRecordQueryAllRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type WithdrawOrderUpdateRequestX struct {
	List []ChannelWithdraw `json:"list"`
	OrderX
}

type OrderActionX struct {
	OrderAction
	CreatedAt time.Time
}

type OrderChannelsX struct {
	OrderChannels
	CreatedAt time.Time
	UpdatedAt time.Time
}
type OrderFeeProfitX struct {
	OrderFeeProfit
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ReceiptRecordX struct {
	ReceiptRecord
	TransAt            JsonTime `json:"transAt, optional"`
	MerchantCallBackAt JsonTime `json:"MerchantCallBackAt"`
	CreatedAt          JsonTime `json:"createdAt, optional"`
}

type IncomeExpenseRecordX struct {
	IncomeExpenseRecord
	CreatedAt JsonTime `json:"createdAt, optional"`
}

type PtBalanceRecordX struct {
	PtBalanceRecord
	CreatedAt JsonTime `json:"createdAt, optional"`
}

type FrozenRecordX struct {
	FrozenRecord
	TransAt   JsonTime `json:"transAt, optional"`
	CreatedAt JsonTime `json:"createdAt"`
	FrozenAt  JsonTime `json:"frozenAt"`
}

type DeductRecordX struct {
	DeductRecord
	TransAt JsonTime `json:"trans_at, optional"`
}

type AllocRecordX struct {
	AllocRecord
	TransAt JsonTime `json:"trans_at, optional"`
}

type ReceiptRecordQueryAllResponseX struct {
	List     []ReceiptRecordX `json:"list"`
	PageNum  int              `json:"pageNum" gorm:"-"`
	PageSize int              `json:"pageSize" gorm:"-"`
	RowCount int64            `json:"rowCount"`
}

type IncomeExpenseQueryResponseX struct {
	List     []IncomeExpenseRecordX `json:"list"`
	PageNum  int                    `json:"pageNum" gorm:"-"`
	PageSize int                    `json:"pageSize" gorm:"-"`
	RowCount int64                  `json:"rowCount"`
}

type PtBalanceRecordsQueryResponseX struct {
	List     []PtBalanceRecordX `json:"list"`
	PageNum  int                `json:"pageNum" gorm:"-"`
	PageSize int                `json:"pageSize" gorm:"-"`
	RowCount int64              `json:"rowCount"`
}
type FrozenRecordQueryAllResponseX struct {
	List     []FrozenRecordX `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type DeductRecordQueryAllResponseX struct {
	List     []DeductRecordX `json:"list"`
	PageNum  int             `json:"pageNum" gorm:"-"`
	PageSize int             `json:"pageSize" gorm:"-"`
	RowCount int64           `json:"rowCount"`
}

type AllocRecordQueryAllResponseX struct {
	List     []AllocRecordX `json:"list"`
	PageNum  int            `json:"pageNum" gorm:"-"`
	PageSize int            `json:"pageSize" gorm:"-"`
	RowCount int64          `json:"rowCount"`
}

type OrderActionQueryAllRequestX struct {
	OrderActionQueryAllRequest
	Language string        `json:"language, optional" gorm:"-"`
	Orders   []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type PersonalRepaymentRequestX struct {
	PersonalRepaymentRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type PersonalRepaymentResponseX struct {
	List     []PersonalRepaymentX `json:"list"`
	PageNum  int                  `json:"pageNum" gorm:"-"`
	PageSize int                  `json:"pageSize" gorm:"-"`
	RowCount int64                `json:"rowCount"`
}

type PersonalRepaymentX struct {
	PersonalRepayment
	TransAt   JsonTime `json:"transAt, optional"`
	CreatedAt JsonTime `json:"createdAt, optional"`
}

type PersonalStatusUpdateResponseX struct {
	PersonalStatusUpdateResponse
	ChannelTransAt JsonTime `json:"channelTransAt, optional"`
}

type OrderFeeProfitQueryAllRequestX struct {
	OrderFeeProfitQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type CalculateProfit struct {
	MerchantCode        string
	OrderNo             string
	Type                string
	CurrencyCode        string
	BalanceType         string
	ChannelCode         string
	ChannelPayTypesCode string
	OrderAmount         float64
}

type PayOrderRequestX struct {
	PayOrderRequest
	MyIp string `json:"my_ip, optional"`
}

type PayQueryRequestX struct {
	PayQueryRequest
	MyIp string `json:"my_ip, optional"`
}

type PayQueryBalanceRequestX struct {
	PayQueryBalanceRequest
	MyIp string `json:"my_ip, optional"`
}

type ProxyPayRequestX struct {
	ProxyPayOrderRequest
	Ip string `json:"ip, optional"`
}

type ProxyPayOrderQueryRequestX struct {
	ProxyPayOrderQueryRequest
	Ip string `json:"ip, optional"`
}

type WithdrawApiOrderRequestX struct {
	WithdrawApiOrderRequest
	MyIp string `json:"my_ip, optional"`
}

type MultipleOrderWithdrawCreateRequestX struct {
	List []OrderWithdrawCreateRequestX `json:"list"`
}

type OrderWithdrawCreateRequestX struct {
	OrderWithdrawCreateRequest
	MerchantCode    string `json:"merchantCode, optional"`
	MerchantOrderNo string `json:"merchant_order_no, optional"`
	UserAccount     string `json:"userAccount, optional"`
	NotifyUrl       string `json:"notify_url, optional"`
	PageUrl         string `json:"page_url, optional"`
	Source          string `json:"source, optional"`
	Type            string `json:"type, optional"`
}

type OrderWithdrawCreateResponse struct {
	OrderX
	Index []string `json:"index"`
	Errs  []string `json:"errs"`
}

type WithdrawApiQueryRequestX struct {
	WithdrawApiQueryRequest
	MyIp string `json:"my_ip, optional"`
}

type OrderLogQueryAllRequestX struct {
	OrderLogQueryAllRequest
	Orders []gormx.Sortx `json:"orders, optional" gorm:"-"`
}

type OrderLogQueryAllResponseX struct {
	List     []TxLog `json:"list"`
	PageNum  int     `json:"pageNum"`
	PageSize int     `json:"pageSize"`
	RowCount int64   `json:"rowCount"`
}

type TransactionLogData struct {
	MerchantCode    string      `json:"merchantCode, optional"`
	MerchantOrderNo string      `json:"merchantOrderNo, optional"`
	OrderNo         string      `json:"orderNo, optional"`
	ChannelOrderNo  string      `json:"channelOrderNo, optional"`
	LogType         string      `json:"logType, optional"`   //交易日志类型(1:錯誤訊息	2:商户请求	3:返回商户错误	4.打给渠道资料	5.渠道返回资料	6.渠道回调资料	7.回调给商户)
	LogSource       string      `json:"logSource, optional"` //日誌來源(1:內充平台、2:支付API、3:代付API、4:代付平台、5:下發API、6:下发平台)
	Content         interface{} `json:"content, optional"`
	ErrCode         string      `json:"errCode, optional"`
	ErrMsg          string      `json:"errMsg, optional"`
	TxOrderSource   string      `json:"txOrderSource, optional"` //1: 平台订单  2: API订单
	TxOrderType     string      `json:"txOrderType, optional"`   //内充 代付 支付 下发
	TraceId         string      `json:"traceId, optional"`
}
