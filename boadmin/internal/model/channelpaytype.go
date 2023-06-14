package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
)

type ChannelPayType struct {
	MyDB  *gorm.DB
	Table string
}

func NewChannelPayType(mydb *gorm.DB, t ...string) *ChannelPayType {
	table := "ch_channel_pay_types"
	if len(t) > 0 {
		table = t[0]
	}
	return &ChannelPayType{
		MyDB:  mydb,
		Table: table,
	}
}

func (m *ChannelPayType) GetByCode(channelPayTypeCode string) (channelPayType *types.ChannelPayType, err error) {
	err = m.MyDB.Table(m.Table).
		Where("code = ? ", channelPayTypeCode).
		Take(&channelPayType).Error
	return
}

func (m *ChannelPayType) CheckChannelPayTypeDuplicated(req types.ChannelPayTypeCreateRequest) (err error) {
	var ChannelPayTypes []types.ChannelPayType
	m.MyDB.Table(m.Table).Where("code = ?", req.Code).Find(&ChannelPayTypes)
	if len(ChannelPayTypes) > 0 {
		return errorz.New(response.CHANNEL_PAYTYPE_DUPLICATED)
	} else {
		return nil
	}
}

func (m *ChannelPayType) SingleInsertChannelPayType(req types.ChannelPayTypeCreateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		return m.MyDB.Table(m.Table).Create(req).Error
	})
}

func (m *ChannelPayType) InsertChannelPayType(req []types.ChannelPayTypeCreateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		return m.MyDB.Table(m.Table).Create(req).Error
	})
}

func (m *ChannelPayType) UpdateChannelPayType(req []types.ChannelPayTypeUpdateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		var IdArr []int64
		for _, m := range req {
			IdArr = append(IdArr, m.ID)
		}
		return m.MyDB.Table(m.Table).Where("id IN ?", IdArr).Updates(&req).Error
	})
}
func (m *ChannelPayType) SingleUpdateChannelPayType(req types.ChannelPayTypeUpdateRequest) (err error) {
	return m.MyDB.Transaction(func(db *gorm.DB) error {
		return m.MyDB.Table(m.Table).Updates(req).Error
	})
}

func (m *ChannelPayType) ChannelPayTypeQueryAll(req types.ChannelPayTypeQueryAllRequestX) (resp *types.ChannelPayTypeQueryAllResponse, err error) {

	var channelPayTypes []types.ChannelPayTypeQueryResponse
	var count int64
	//var terms []string
	db := m.MyDB.Table("ch_channel_pay_types cpt ").
		Joins("left join ch_channels c on cpt.channel_code = c.code ").
		Joins("left join ch_pay_types pt on pt.code = cpt.pay_type_code ")

	selectX :=
		"c.currency_code           	as currency_code," +
			"c.name           			as channel_name," +
			"pt.name	        		as pay_type_name," +
			"cpt.id                		as id," +
			"cpt.code              		as code," +
			"cpt.channel_code      		as channel_code," +
			"cpt.pay_type_code     		as pay_type_code," +
			"cpt.fee      		   		as fee," +
			"cpt.handling_fee      		as handling_fee," +
			"cpt.max_internal_charge    as max_internal_charge," +
			"cpt.daily_tx_limit      	as daily_tx_limit," +
			"cpt.single_min_charge      as single_min_charge," +
			"cpt.single_max_charge      as single_max_charge," +
			"cpt.fixed_amount      		as fixed_amount," +
			"cpt.bill_date      		as bill_date," +
			"cpt.status      			as status," +
			"cpt.is_proxy				as is_proxy," +
			"cpt.device   				as device "

	if len(req.ChannelCode) > 0 {
		db.Where("cpt.channel_code like ?", "%"+req.ChannelCode+"%")
		//terms = append(terms, fmt.Sprintf("cpt.channel_code like '%%%s%%'", req.ChannelCode))
	}
	if len(req.ChannelName) > 0 {
		db.Where("c.name like ?", "%"+req.ChannelName+"%")
		//terms = append(terms, fmt.Sprintf("c.name like '%%%s%%'", req.ChannelName))
	}
	if len(req.Currency) > 0 {
		db.Where("c.currency_code = ?", req.Currency)
		//terms = append(terms, fmt.Sprintf("c.currency_code = '%s'", req.Currency))
	}
	if len(req.PayTypeName) > 0 {
		db.Where("pt.name like ?", "%"+req.PayTypeName+"%")
		//terms = append(terms, fmt.Sprintf("pt.name like '%%%s%%'", req.PayTypeName))
	}
	if len(req.Status) > 0 {
		db.Where("cpt.status = ?", req.Status)
		//terms = append(terms, fmt.Sprintf("cpt.status = '%s'", req.Status))
	}
	if len(req.IsProxy) > 0 {
		db.Where("cpt.is_proxy = ?", req.IsProxy)
		//terms = append(terms, fmt.Sprintf("cpt.is_proxy = '%s'", req.IsProxy))
	}
	//term := strings.Join(terms, " AND ")

	if err = db.Select(selectX).
		Scopes(gormx.Paginate(req)).
		//Clauses(clause.OrderBy{
		//	Expression: clause.Expr{SQL: "FIELD(cpt.status,?)", Vars: []interface{}{[]string{"1", "2", "0"}}, WithoutParentheses: true},
		//}).
		Order("FIELD(cpt.`STATUS`,'1','2','0')").
		Scopes(gormx.Sort(req.Orders)).
		Find(&channelPayTypes).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.
		Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.ChannelPayTypeQueryAllResponse{
		List:     channelPayTypes,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
