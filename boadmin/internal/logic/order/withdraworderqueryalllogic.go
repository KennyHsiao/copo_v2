package order

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"gorm.io/gorm"
	"sort"
	"strings"
	"sync"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WithdrawOrderQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawOrderQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawOrderQueryAllLogic {
	return WithdrawOrderQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type totalWithDrawOrder []types.WithDrawOrder

func (p totalWithDrawOrder) Len() int {
	return len(p)
}

func (p totalWithDrawOrder) Less(i, j int) bool {
	created1, _ := time.Parse("2006-01-02 15:04:05", p[i].CreatedAt)
	created2, _ := time.Parse("2006-01-02 15:04:05", p[j].CreatedAt)
	return created1.Before(created2)
}

func (p totalWithDrawOrder) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (l *WithdrawOrderQueryAllLogic) WithdrawOrderQueryAll(req types.WithDrawOrderQueryAllRequest) (resp *types.WithdrawOrderQueryAllResponse, err error) {
	var totalWithDrawOrders []types.WithDrawOrder
	var currencies []types.Currency

	if err = l.svcCtx.MyDB.Table("bs_currencies").Find(&currencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(len(currencies))
	for i, _ := range currencies {
		go func(i int) error {
			defer wg.Done()
			withdrawOrders, errW := withdrawReviewOrderQueryAll(l.svcCtx.MyDB, req, currencies[i].Code)
			if errW != nil {
				return errW
			}
			if len(withdrawOrders) > 0 {
				totalWithDrawOrders = append(totalWithDrawOrders, withdrawOrders...)
			}
			return nil
		}(i)

	}
	wg.Wait()

	sort.Sort(totalWithDrawOrder(totalWithDrawOrders))

	count := len(totalWithDrawOrders)

	resp = &types.WithdrawOrderQueryAllResponse{
		List:     totalWithDrawOrders,
		RowCount: int64(count),
	}
	return
}

func withdrawReviewOrderQueryAll(db *gorm.DB, req types.WithDrawOrderQueryAllRequest, currency string) (resp []types.WithDrawOrder, err error) {
	var withDrawOrders []types.WithDrawOrder
	var terms []string

	if len(req.MerchantCode) > 0 {
		terms = append(terms, fmt.Sprintf("`merchant_code` = '%s'", req.MerchantCode))
	}
	if len(req.StartAt) > 0 {
		terms = append(terms, fmt.Sprintf("`created_at` >= '%s'", req.StartAt))
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		terms = append(terms, fmt.Sprintf("`created_at` < '%s'", endAt))
	}
	terms = append(terms, fmt.Sprintf("`currency_code` = '%s'", currency))
	terms = append(terms, fmt.Sprintf("`type` = 'XF'"))
	terms = append(terms, fmt.Sprintf("`status` IN ('0','1')"))
	term := strings.Join(terms, "AND")
	err = db.Table("tx_orders").Where(term).Find(&withDrawOrders).Error

	if len(withDrawOrders) > 0 {
		for i, _ := range withDrawOrders {
			withDrawOrders[i].CreatedAt = utils.ParseTime(withDrawOrders[i].CreatedAt)
			for strings.HasSuffix(withDrawOrders[i].OrderAmount, "0") {
				withDrawOrders[i].OrderAmount = strings.TrimSuffix(withDrawOrders[i].OrderAmount, "0")
			}
			if strings.HasSuffix(withDrawOrders[i].OrderAmount, ".") {
				withDrawOrders[i].OrderAmount = strings.TrimSuffix(withDrawOrders[i].OrderAmount, ".")
			}
		}
	}

	return withDrawOrders, nil
}
