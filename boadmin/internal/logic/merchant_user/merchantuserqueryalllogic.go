package merchant_user

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantUserQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantUserQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantUserQueryAllLogic {
	return MerchantUserQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantUserQueryAllLogic) MerchantUserQueryAll(req types.MerchantUserQueryAllRequestX) (resp *types.MerchantUserQueryAllResponse, err error) {
	var merchantUsers []types.MerchantUser
	//var terms []string
	var count int64

	db := l.svcCtx.MyDB.Table("au_user_merchants as aum").
		Joins("join au_users au on aum.user_account = au.account").
		Joins("join au_user_roles aur on au.id = aur.user_id").
		Joins("join au_roles ar on aur.role_id = ar.id")

	selectX := " au.id," +
		" aum.merchant_code," +
		" au.account," +
		" au.status," +
		" au.is_freeze," +
		" au.disable_delete," +
		" au.registered_at AS registered_at," +
		" au.last_login_at AS last_login_at," +
		" ar.name as role_names, " +
		" ar.id as role_id "

	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf("aum.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("aum.merchant_code = ?", req.MerchantCode)
	}
	if len(req.Status) > 0 {
		//terms = append(terms, fmt.Sprintf("status = '%s'", req.Status))
		db = db.Where("status = ?", req.Status)
	}
	if len(req.Account) > 0 {
		//terms = append(terms, fmt.Sprintf("account = '%s'", req.Account))
		db = db.Where("account = ?", req.Account)
	}
	if req.RoleId > 0 {
		//terms = append(terms, fmt.Sprintf("ar.id = '%d' ", req.RoleId))
		db = db.Where("ar.id = ?", req.RoleId)

	}

	//term := strings.Join(terms, " AND ")

	myDB := db.Select(selectX)
	//Group("aum.user_account, aum.merchant_code").
	//Where(term)

	if err = myDB.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = myDB.
		Scopes(gormx.Sort(req.Orders)).
		Scopes(gormx.Paginate(req)).
		Find(&merchantUsers).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantUserQueryAllResponse{
		List:     merchantUsers,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
