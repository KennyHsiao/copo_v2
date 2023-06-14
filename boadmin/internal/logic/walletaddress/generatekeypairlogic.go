package walletaddress

import (
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateKeyPairLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateKeyPairLogic(ctx context.Context, svcCtx *svc.ServiceContext) GenerateKeyPairLogic {
	return GenerateKeyPairLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateKeyPairLogic) GenerateKeyPair(req *types.GenerateKeyPairRequest) (resp *types.GenerateKeyPairResponse, err error) {

	return
}
