package model

import (
	"com.copo/bo_service/boadmin/internal/config"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type PayType struct {
	Config config.Config
	svcCtx *svc.ServiceContext
	MyDB   *gorm.DB
	Table  string
}

func NewPayType(mydb *gorm.DB, t ...string) *PayType {
	table := "ch_pay_types"
	if len(t) > 0 {
		table = t[0]
	}
	return &PayType{
		MyDB:  mydb,
		Table: table,
	}
}

func (c *PayType) PayTypeQueryAll(req types.PayTypeQueryAllRequestX) (resp *types.PayTypeQueryAllResponse, err error) {
	var currencies []types.PayType
	var count int64
	//var terms []string
	db := c.MyDB
	db2 := c.MyDB
	if len(req.Code) > 0 {
		//terms = append(terms, fmt.Sprintf("`code` like '%%%s%%'", req.Code))
		db = db.Where("code like ?", "%"+req.Code+"%")
		db2 = db2.Where("code like ?", "%"+req.Code+"%")
	}
	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name like '%%%s%%'", req.Name))
		db = db.Where("name like ?", "%"+req.Name+"%")
		db2 = db2.Where("name like ?", "%"+req.Name+"%")
	}
	if len(req.Currency) > 0 {
		//terms = append(terms, fmt.Sprintf("currency like '%%%s%%'", req.Currency))
		db = db.Where("currency like ?", "%"+req.Currency+"%")
		db2 = db2.Where("currency like ?", "%"+req.Currency+"%")
	}

	//term := strings.Join(terms, " AND ")
	if err = db2.Table(c.Table).Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table(c.Table).
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		//Order("CONVERT(name USING GBK)").
		Find(&currencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.PayTypeQueryAllResponse{
		List:     currencies,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}

func (p *PayType) CheckPayTypeDuplicated(req types.PayTypeCreateRequest) bool {
	var payTypes []types.PayType
	p.MyDB.Table(p.Table).Where("code = ?", req.Code).Find(&payTypes)
	if len(payTypes) > 0 {
		return true
	} else {
		return false
	}
}

func (p *PayType) CheckPayTypeExist(req types.PayTypeUpdateRequest) (err error) {
	var payTypes []types.PayType
	p.MyDB.Table(p.Table).Where("code = ?", req.Code).Find(&payTypes)
	if len(payTypes) > 0 {
		return nil
	} else {
		return errorz.New(response.PAY_TYPE_DUPLICATED, err.Error())
	}
}
func (p *PayType) GenerateNewPayTypeNum(payType string) (payTypeNum string) {
	var codeNumList string
	row := p.MyDB.Table(p.Table).Where("code = ?", payType).Select("code_num").Row()
	row.Scan(&codeNumList)
	if codeNumList == "" {
		return payType + "1"
	} else {
		codeNumArr := strings.Split(codeNumList, ",")
		lastOne := codeNumArr[len(codeNumArr)-1]
		num, _ := strconv.Atoi(string(lastOne[len(lastOne)]))
		return payType + fmt.Sprintf("%d", num+1)
	}
}
