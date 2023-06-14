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

type TimezonexDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimezonexDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) TimezonexDeleteLogic {
	return TimezonexDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimezonexDeleteLogic) TimezonexDelete(req types.TimezonexDeleteRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		if err = l.svcCtx.MyDB.Table("bs_timezone").Delete(&req).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return
	})
}
