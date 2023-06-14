package channeldata

import (
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/utils"
	"context"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelDataQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataQueryAllLogic {
	return ChannelDataQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataQueryAllLogic) ChannelDataQueryAll(req types.ChannelDataQueryAllRequestX) (resp *types.ChannelDataQueryAllResponseX, err error) {
	channelPayTypeList := []types.ChannelPayType{}
	var channels []types.ChannelData2
	var count int64
	db := l.svcCtx.MyDB

	if len(req.Code) > 0 {
		db = db.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		db = db.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("currency_code = ?", req.CurrencyCode)
	}
	if len(req.Status) > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if len(req.IsProxy) > 0 {
		db = db.Where("is_proxy = ?", req.IsProxy)
	} else {
		db = db.Where("(is_proxy = '1' OR is_proxy = '0')")
	}
	if len(req.Device) > 0 {
		if strings.EqualFold("All", req.Device) {
			db = db.Where("device = ?", req.Device)
		} else {
			db = db.Where("(device = ? OR device = 'All')", req.Device)
		}
	}
	db = db.Where("status != '0'")
	db.Table("ch_channels").Count(&count)
	//Preload("Banks")
	err = db.Table("ch_channels").
		Preload("Banks").
		Preload("ChannelPayTypeList").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&channels).Error

	//var payTypeMapList []types.PayTypeMap
	channelist := []types.ChannelData2{}
	if len(channels) > 0 {
		for _, channel := range channels {
			err = l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("channel_code=?", channel.Code).Find(&channelPayTypeList).Error
			channel.ChannelPayTypeList = channelPayTypeList
			channel.CreatedAt = utils.ParseTime(channel.CreatedAt)
			channel.UpdatedAt = utils.ParseTime(channel.UpdatedAt)
			channelist = append(channelist, channel)
		}
	}
	resp = &types.ChannelDataQueryAllResponseX{
		List:     channelist,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return resp, err
}

type ChannelDa struct {
	ID                      int64   `json:"id, optional"`
	Code                    string  `json:"code, optional"`
	Name                    string  `json:"name, optional"`
	ProjectName             string  `json:"projectName, optional"`
	IsProxy                 string  `json:"isProxy, optional"`
	IsNzPre                 string  `json:"isNzPre, optional"`
	ApiUrl                  string  `json:"apiUrl, optional"`
	CurrencyCode            string  `json:"currencyCode, optional"`
	ChannelWithdrawCharge   float64 `json:"channelWithdrawCharge, optional"`
	Balance                 float64 `json:"balance, optional"`
	Status                  string  `json:"status, optional"`
	Device                  string  `json:"device,optional"`
	MerId                   string  `json:"merId, optional"`
	MerKey                  string  `json:"merKey, optional"`
	PayUrl                  string  `json:"payUrl, optional"`
	PayQueryUrl             string  `json:"payQueryUrl, optional"`
	PayQueryBalanceUrl      string  `json:"payQueryBalanceUrl, optional"`
	ProxyPayUrl             string  `json:"proxyPayUrl, optional"`
	ProxyPayQueryUrl        string  `json:"proxyPayQueryUrl, optional"`
	ProxyPayQueryBalanceUrl string  `json:"proxyPayQueryBalanceUrl, optional"`
	WhiteList               string  `json:"whiteList, optional"`
	//PayTypeMapList          []PayTypeMap     `json:"payTypeMapList, optional" gorm:"-"`
	PayTypeMap string `json:"payTypeMap, optional"`
	//ChannelPayTypeList      []ChannelPayType `json:"channelPayTypeList, optional" gorm:"foreignKey:ChannelCode;references:Code"`
	ChannelPort     string  `json:"channelPort, optional"`
	WithdrawBalance float64 `json:"withdrawBalance, optional"`
	ProxypayBalance float64 `json:"proxypayBalance, optional"`
	//BankCodeMapList         []BankCodeMap    `json:"bankCodeMapList, optional" gorm:"-"`
	//Banks                   []Bank           `json:"banks, optional" gorm:"many2many:ch_channel_banks;foreignKey:Code;joinForeignKey:channel_code;references:bank_no;joinReferences:bank_no"`
}
