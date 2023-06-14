package bankblockaccount

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankBlockAccountQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankBlockAccountQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankBlockAccountQueryAllLogic {
	return BankBlockAccountQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankBlockAccountQueryAllLogic) BankBlockAccountQueryAll(req types.BankBlockAccountQueryAllRequest) (resp *types.BankBlockAccountQueryAllResponse, err error) {
	var bankBlockAccounts []types.BankBlockAccount
	var count int64
	db := l.svcCtx.MyDB

	if len(req.BankAccount) > 0 {
		db = db.Where("`bank_account` like ?", "%"+req.BankAccount+"%")
	}
	if len(req.Name) > 0 {
		db = db.Where("`name` like ?", "%"+req.Name+"%")
	}
	if len(req.StartAt) > 0 {
		db = db.Where("`created_at` >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		db = db.Where("`created_at` < ?", endAt)
	}
	db.Table("bk_block_account").Count(&count)
	err = db.Table("bk_block_account").Scopes(gormx.Paginate(req)).Order("created_at desc").Find(&bankBlockAccounts).Error

	bankBlockAccounts = changeDateValue(bankBlockAccounts)

	resp = &types.BankBlockAccountQueryAllResponse{
		List:     bankBlockAccounts,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}

func changeDateValue(b []types.BankBlockAccount) []types.BankBlockAccount {
	for i, _ := range b {
		b[i].CreatedAt = utils.ParseTime(b[i].CreatedAt)
	}
	return b
}
