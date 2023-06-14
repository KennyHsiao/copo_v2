package timezonex

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimezonexUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimezonexUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) TimezonexUpdateLogic {
	return TimezonexUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimezonexUpdateLogic) TimezonexUpdate(req types.TimezonexUpdateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		timezone := &types.TimezonexUpdate{
			TimezonexUpdateRequest: req,
		}
		if err = l.svcCtx.MyDB.Table("bs_timezone").Updates(timezone).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return
	})
}
