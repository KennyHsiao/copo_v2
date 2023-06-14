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

type InternalChargeOrderQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInternalChargeOrderQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) InternalChargeOrderQueryAllLogic {
	return InternalChargeOrderQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

type totalInterChargeOrder []types.InternalChargeOrder

func (p totalInterChargeOrder) Len() int {
	return len(p)
}

func (p totalInterChargeOrder) Less(i, j int) bool {
	created1, _ := time.Parse("2006-01-02 15:04:05", p[i].CreatedAt)
	created2, _ := time.Parse("2006-01-02 15:04:05", p[j].CreatedAt)
	return created1.Before(created2)
}

func (p totalInterChargeOrder) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (l *InternalChargeOrderQueryAllLogic) InternalChargeOrderQueryAll(req types.InternalChargeOrderQueryAllRequest) (resp *types.InternalChargeOrderQueryAllResponse, err error) {
	var totalInterChargeOrders []types.InternalChargeOrder
	var currencies []types.Currency

	if err = l.svcCtx.MyDB.Table("bs_currencies").Find(&currencies).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	var wg sync.WaitGroup
	wg.Add(len(currencies))
	for i, _ := range currencies {
		go func(i int) error {
			defer wg.Done()
			internalChargeOrders, errI := internalChargeOrderQueryAll(l.svcCtx.MyDB, req, l.svcCtx, currencies[i].Code)
			if errI != nil {
				return errI
			}
			if len(internalChargeOrders) > 0 {
				totalInterChargeOrders = append(totalInterChargeOrders, internalChargeOrders...)
			}
			return nil
		}(i)
	}
	wg.Wait()

	sort.Sort(totalInterChargeOrder(totalInterChargeOrders))

	count := len(totalInterChargeOrders)

	resp = &types.InternalChargeOrderQueryAllResponse{
		List:     totalInterChargeOrders,
		RowCount: int64(count),
	}
	return
}

func internalChargeOrderQueryAll(db *gorm.DB, req types.InternalChargeOrderQueryAllRequest, svcCtx *svc.ServiceContext, currency string) (resp []types.InternalChargeOrder, err error) {
	var interChargeOrders []types.InternalChargeOrder
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
	terms = append(terms, fmt.Sprintf("`type` = 'NC'"))
	terms = append(terms, fmt.Sprintf("`status` = '%s'", req.Status))
	terms = append(terms, fmt.Sprintf("`currency_code` = '%s'", currency))
	term := strings.Join(terms, "AND")
	if err = db.Table("tx_orders").Where(term).Find(&interChargeOrders).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, _ := range interChargeOrders {
		interChargeOrders[i].CreatedAt = utils.ParseTime(interChargeOrders[i].CreatedAt)
		if len(interChargeOrders[i].InternalChargeOrderPath) > 0 {
			interChargeOrders[i].InternalChargeOrderPath =
				svcCtx.Config.ResourceHost + interChargeOrders[i].InternalChargeOrderPath
		}
		for strings.HasSuffix(interChargeOrders[i].OrderAmount, "0") {
			interChargeOrders[i].OrderAmount = strings.TrimSuffix(interChargeOrders[i].OrderAmount, "0")
		}
		if strings.HasSuffix(interChargeOrders[i].OrderAmount, ".") {
			interChargeOrders[i].OrderAmount = strings.TrimSuffix(interChargeOrders[i].OrderAmount, ".")
		}
	}

	return interChargeOrders, nil
}
