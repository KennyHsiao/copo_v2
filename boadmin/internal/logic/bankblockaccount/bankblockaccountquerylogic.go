package bankblockaccount

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type BankBlockAccountQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankBlockAccountQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankBlockAccountQueryLogic {
	return BankBlockAccountQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankBlockAccountQueryLogic) BankBlockAccountQuery(req types.BankBlockAccountQueryRequest) (resp *types.BankBlockAccountQueryResponse, err error) {

	err = l.svcCtx.MyDB.Table("bk_block_account").First(&resp, req.ID).Error
	resp.CreatedAt = utils.ParseTime(resp.CreatedAt)

	return
}
