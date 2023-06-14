package orderaction

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/gioco-play/easy-i18n/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderActionQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderActionQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderActionQueryAllLogic {
	return OrderActionQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderActionQueryAllLogic) OrderActionQueryAll(req types.OrderActionQueryAllRequestX) (resp *types.OrderActionQueryAllResponse, err error) {
	var orderActions []types.OrderAction
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	if len(req.OrderNo) > 0 {
		db = db.Where("order_no = ?", req.OrderNo)
		//terms = append(terms, fmt.Sprintf(" order_no = '%s'", req.OrderNo))
	}
	if len(req.UserAccount) > 0 {
		db = db.Where("user_account = ?", req.UserAccount)
		//terms = append(terms, fmt.Sprintf(" user_account = '%s'", req.UserAccount))
	}

	if len(req.StartAt) > 0 {
		db = db.Where("created_at >= ?", req.StartAt)
		//terms = append(terms, fmt.Sprintf(" created_at >= '%s'", req.StartAt))
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		db = db.Where("created_at < ?", endAt)
		//terms = append(terms, fmt.Sprintf(" created_at < '%s'", endAt))
	}

	//term := strings.Join(terms, "AND")
	db.Table("tx_order_actions").Count(&count)
	if err = db.Table(" tx_order_actions").
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).
		Find(&orderActions).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	utils.SetI18n(req.Language)
	for i, action := range orderActions {
		orderActions[i].CreatedAt = utils.ParseTime(action.CreatedAt)
		orderActions[i].Action = i18n.Sprintf(action.Action)
		logx.Infof(orderActions[i].Action)
	}

	resp = &types.OrderActionQueryAllResponse{
		List:     orderActions,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
