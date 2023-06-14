package currency

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type CurrencyBatchSortLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCurrencyBatchSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) CurrencyBatchSortLogic {
	return CurrencyBatchSortLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CurrencyBatchSortLogic) CurrencyBatchSort(req *types.CurrencyBatchSortRequest) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		for index, id := range req.List {
			if err := tx.Updates(&types.Currency{
				ID:        id,
				SortOrder: int64(index + 1),
			}).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		return nil
	})
}
