package model

import (
	"com.copo/bo_service/boadmin/internal/types"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type Bank struct {
	MyDB  *gorm.DB
	Table string
}

func GetNewBankCode(db *gorm.DB, CurrencyCode string) string {
	var code string
	row := db.Table("bk_banks").Select("max(bank_no)").Where("currency_code = ?", CurrencyCode).Row()
	row.Scan(&code)
	codeNum, _ := strconv.Atoi(code)
	return fmt.Sprintf("%d", codeNum+1)
}

func CheckBankIsExist(db *gorm.DB, req types.BankUpdateRequest) error {
	var bank = &types.Bank{}
	db.Table("bk_banks").Where("bank_no = ? AND currency_code", req.BankNo, req.CurrencyCode).Find(&bank)
	if bank == nil {
		logx.Error("查无银行资料")
		return errors.New("查无银行资料")
	}
	return nil
}

func CheckBankIsDuplicated(db *gorm.DB, req types.BankCreateRequest) (isDuplicated bool, bank *types.Bank) {
	var banks []types.Bank
	var terms []string
	var values []interface{}

	if len(req.BankNo) > 0 {
		terms = append(terms, "bank_no = ?")
		values = append(values, req.BankNo)
	}
	if len(req.BankName) > 0 {
		terms = append(terms, "bank_name = ?")
		values = append(values, req.BankName)
	}
	if len(req.BankNameEn) > 0 {
		terms = append(terms, "bank_name_en = ?")
		values = append(values, req.BankNameEn)
	}
	if len(req.Abbr) > 0 {
		terms = append(terms, "abbr = ?")
		values = append(values, req.Abbr)
	}
	clause := strings.Join(terms, " OR ")
	db.Table("bk_banks").Where("currency_code = ?", req.CurrencyCode).Where(clause, values...).Find(&banks)

	if len(banks) > 0 {
		logx.Info("新增银行资料重复:", banks)
		return true, &banks[0]
	} else {
		return false, nil
	}

}

func CheckBankIsDuplicatedUp(db *gorm.DB, req types.BankUpdateRequest) (isDuplicated bool) {
	var banks []types.Bank
	var terms []string
	var values []interface{}

	if len(req.BankName) > 0 {
		terms = append(terms, "bank_name = ?")
		values = append(values, req.BankName)
	}
	if len(req.BankNameEn) > 0 {
		terms = append(terms, "bank_name_en = ?")
		values = append(values, req.BankNameEn)
	}
	if len(req.Abbr) > 0 {
		terms = append(terms, "abbr = ?")
		values = append(values, req.Abbr)
	}
	clause := strings.Join(terms, " OR ")
	db.Table("bk_banks").Where("currency_code = ?", req.CurrencyCode).Where(clause, values...).Find(&banks)

	if len(banks) > 1 {
		logx.Info("更新银行资料重复:", banks)
		return true
	} else {
		return false
	}

}
