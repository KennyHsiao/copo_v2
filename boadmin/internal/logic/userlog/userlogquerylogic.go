package userlog

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserLogQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserLogQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserLogQueryLogic {
	return UserLogQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserLogQueryLogic) UserLogQuery(req *types.UserLogQueryAllRequestX) (resp *types.UserLogQueryAllResponseX, err error) {

	var count int64
	var userLogs []types.UserLogX
	//var terms []string
	db := l.svcCtx.MyDB
	db2 := l.svcCtx.MyDB

	if len(req.Operating) > 0 {
		//terms = append(terms, fmt.Sprintf("operating like '%%%s%%'", req.Operating))
		db = db.Where("Operating like ?", "%"+req.Operating+"%")
		db2 = db2.Where("Operating like ?", "%"+req.Operating+"%")
	}

	if len(req.AccountName) > 0 {
		//terms = append(terms, fmt.Sprintf("account_name like '%%%s%%'", req.AccountName))
		db = db.Where("account_name like ?", "%"+req.AccountName+"%")
		db2 = db2.Where("account_name like ?", "%"+req.AccountName+"%")
	}

	if len(req.RequestUnit) > 0 {
		//terms = append(terms, fmt.Sprintf("request_unit like '%%%s%%'", req.RequestUnit))
		db = db.Where("request_unit like ?", "%"+req.RequestUnit+"%")
		db2 = db2.Where("request_unit like ?", "%"+req.RequestUnit+"%")
	}

	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("type = '%s'", req.Type))
		db = db.Where("type = ?", req.Type)
		db2 = db2.Where("type = ?", req.Type)
	}

	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf("`created_at` >= '%s'", req.StartAt))
		db = db.Where("created_at >= ?", req.StartAt)
		db2 = db2.Where("created_at >= ?", req.StartAt)
	}

	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf("`created_at` < '%s'", endAt))
		db = db.Where("created_at < ?", endAt)
		db2 = db2.Where("created_at < ?", endAt)
	}

	//term := strings.Join(terms, " AND ")
	db2.Table("au_user_log").Count(&count)
	if err = db.Table("au_user_log").
		Scopes(gormx.Paginate(*req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&userLogs).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return &types.UserLogQueryAllResponseX{
		List:     userLogs,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}, nil
}
