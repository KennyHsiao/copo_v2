package paytype

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeBatchSortLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeBatchSortLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeBatchSortLogic {
	return PayTypeBatchSortLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeBatchSortLogic) PayTypeBatchSort(req *types.BatchSortRequest) error {

	for index, id := range req.List {
		sortNum := fmt.Sprintf("%03d", index)
		if err := l.svcCtx.MyDB.Updates(&types.PayType{
			ID:      id,
			SortNum: sortNum,
		}).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	return nil
}
