package banks

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankQueryLogic {
	return BankQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankQueryLogic) BankQuery(req types.BankQueryRequest) (resp *types.BankQueryResponse, err error) {
	if err = l.svcCtx.MyDB.Table("bk_banks").Take(&resp, req.ID).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
