package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"gorm.io/gorm"
)

type Currency struct {
	MyDB  *gorm.DB
	Table string
}

func NewCurrency(mydb *gorm.DB, t ...string) *Currency {
	table := "bs_currencies"
	if len(t) > 0 {
		table = t[0]
	}
	return &Currency{
		MyDB:  mydb,
		Table: table,
	}
}

func (c *Currency) CurrencyQueryAll(req types.CurrencyQueryAllRequestX) (resp *types.CurrencyQueryAllResponse, err error) {
	var currencies []types.Currency
	var count int64
	//var terms []string
	db := c.MyDB.Table("bs_currencies")

	if len(req.Code) > 0 {
		//terms = append(terms, fmt.Sprintf("`code` like '%%%s%%'", req.Code))
		db = db.Where("`code` LIKE ?", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name like '%%%s%%'", req.Name))
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	//term := strings.Join(terms, " AND ")

	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&currencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.CurrencyQueryAllResponse{
		List:     currencies,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
