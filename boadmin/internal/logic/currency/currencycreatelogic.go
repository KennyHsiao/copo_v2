package currency

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyCreateLogic {
	return CurrencyCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyCreateLogic) CurrencyCreate(req types.CurrencyCreateRequest) (err error) {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		data := struct {
			SortOrderMax string `json:"sort_order_max"`
		}{}
		if err := tx.Table("bs_currencies").Select("MAX(sort_order) as sort_order_max").Take(&data).Error; err != nil {
			return errorz.New(response.CREATE_FAILURE, err.Error())
		}

		currency := &types.CurrencyCreate{
			CurrencyCreateRequest: req,
		}
		sortOrder, _ := strconv.ParseInt(data.SortOrderMax, 10, 64)
		currency.SortOrder = sortOrder + 1
		if err := tx.Table("bs_currencies").Create(currency).Error; err != nil {
			return errorz.New(response.CREATE_FAILURE, err.Error())
		}

		return nil
	})

}
