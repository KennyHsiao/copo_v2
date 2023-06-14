package banks

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankQueryAllLogic {
	return BankQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankQueryAllLogic) BankQueryAll(req types.BankQueryAllRequestX) (resp *types.BankQueryAllResponse, err error) {
	var banks []types.Bank
	var count int64
	db := l.svcCtx.MyDB

	if len(req.CurrencyCode) > 0 {
		db = db.Where("currency_code = ?", req.CurrencyCode)
	}
	if len(req.BankName) > 0 {
		db = db.Where("bank_name like ?", "%"+req.BankName+"%")
	}
	if len(req.BankNo) > 0 {
		db = db.Where("bank_no = ?", req.BankNo)
	}

	db.Table("bk_banks").Count(&count)
	err = db.Table("bk_banks").Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&banks).Error
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.BankQueryAllResponse{
		List:     banks,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return resp, err
}
