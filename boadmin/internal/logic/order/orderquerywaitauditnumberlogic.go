package order

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

type OrderQueryWaitAuditNumberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderQueryWaitAuditNumberLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderQueryWaitAuditNumberLogic {
	return OrderQueryWaitAuditNumberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderQueryWaitAuditNumberLogic) OrderQueryWaitAuditNumber(req types.OrderQueryWaitAuditNumberRequest) (resp *types.OrderQueryWaitAuditNumberResponse, err error) {
	var waitAuditNumber int64
	var terms []string

	if len(req.Type) > 0 {
		terms = append(terms, fmt.Sprintf("`type` = '%s'", req.Type))
		if req.Type == constants.ORDER_TYPE_NC {
			terms = append(terms, fmt.Sprintf("`status` = '1'"))
		}
		if req.Type == constants.ORDER_TYPE_XF {
			terms = append(terms, fmt.Sprintf("`status` IN ('0','1')"))
		}
	}
	if len(req.StartAt) > 0 {
		terms = append(terms, fmt.Sprintf("`created_at` >= '%s'", req.StartAt))
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		terms = append(terms, fmt.Sprintf("`created_at` < '%s'", endAt))
	}

	term := strings.Join(terms, "AND")

	if err = l.svcCtx.MyDB.Table("tx_orders").Where(term).Count(&waitAuditNumber).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.OrderQueryWaitAuditNumberResponse{
		WaitAuditNumber: waitAuditNumber,
	}
	return
}
