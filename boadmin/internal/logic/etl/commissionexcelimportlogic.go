package etl

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommissionExcelImportLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommissionExcelImportLogic(ctx context.Context, svcCtx *svc.ServiceContext) CommissionExcelImportLogic {
	return CommissionExcelImportLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommissionExcelImportLogic) CommissionExcelImport(req *types.UploadExcelRequestX) error {
	f, err := excelize.OpenReader(req.UploadFile)

	fmt.Println(req.UploadFile)
	if err != nil {
		fmt.Println(err)
		return errorz.New(response.SYSTEM_ERROR)
	}

	rows, err := f.GetRows(f.GetSheetName(f.GetActiveSheetIndex()))
	if err != nil {
		fmt.Println(err)
		return errorz.New(response.SYSTEM_ERROR)
	}
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		//var merchantBalances []types.MerchantBalance
		//if err := l.svcCtx.MyDB.Table("mc_merchant_balances").Find(&merchantBalances).Error; err != nil {
		//	return errorz.New(response.DATABASE_FAILURE, err.Error())
		//}
		//merchantBalanceMap := make(map[string]types.MerchantBalance)
		//for _, balance := range merchantBalances {
		//	merchantBalanceMap[balance.MerchantCode] = balance
		//}
		for i, row := range rows {
			if i == 0 {
				continue
			}
			merchantCode := ""
			currencyCode := ""
			money := 0.0
			var merchantYJB types.MerchantBalance
			for k, colCell := range row {
				if k == 0 {
					currencyCode = colCell
				}
				if k == 1 {
					merchantCode = colCell
				}
				if k == 3 {
					money, _ = strconv.ParseFloat(colCell, 64)
				}
				merchantYJB.Balance = money

			}
			if err := tx.Table("mc_merchant_balances").
				Where("merchant_code = ?", merchantCode).
				Where("currency_code = ?", currencyCode).
				Where("balance_type = ?", constants.YJ_BALANCE).Updates(merchantYJB).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}
		//
		//for _, row := range rows {
		//	for _, colCell := range row {
		//		fmt.Println(colCell, "\t")
		//	}
		//	fmt.Println()
		//}

		return nil
	})

}
