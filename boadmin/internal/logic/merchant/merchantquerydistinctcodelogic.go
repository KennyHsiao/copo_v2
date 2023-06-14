package merchant

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantQueryDistinctCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantQueryDistinctCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantQueryDistinctCodeLogic {
	return MerchantQueryDistinctCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantQueryDistinctCodeLogic) MerchantQueryDistinctCode(req types.MerchantQueryDistinctCode) (resp []string, err error) {
	var terms []string

	if req.CurrencyCode != "" {
		terms = append(terms, fmt.Sprintf("mc.currency_code = '%s'", req.CurrencyCode))
	}

	if req.Status != "" {
		terms = append(terms, fmt.Sprintf("mc.status = '%s'", req.Status))
	} else {
		terms = append(terms, "mc.status = '1'")
	}

	if req.AgentStatus != "" {
		terms = append(terms, fmt.Sprintf("mer.agent_status = '%s'", req.AgentStatus))
	}

	term := strings.Join(terms, " AND ")

	if err = l.svcCtx.MyDB.Table("mc_merchant_currencies as mc").
		Joins("join mc_merchants mer on mer.code = mc.merchant_code").
		Order("merchant_code asc").
		Where(term).
		Distinct().
		Pluck("mc.merchant_code", &resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}
