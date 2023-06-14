package commissionMonthReport

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MonthReportQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMonthReportQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MonthReportQueryAllLogic {
	return MonthReportQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MonthReportQueryAllLogic) MonthReportQueryAll(req *types.CommissionMonthReportQueryAllRequestX) (resp *types.CommissionMonthReportQueryAllResponseX, err error) {
	var reports []types.CommissionMonthReportX
	var count int64
	//var terms []string
	var merchant *types.Merchant
	var totalCommissionAmount float64
	db := l.svcCtx.MyDB.Table("cm_commission_month_reports")

	// JWT 為商戶帳號時 只能取子孫代理(包含自己)
	jwtMerchantCode := l.ctx.Value("merchantCode").(string)
	if len(jwtMerchantCode) > 0 {
		ux := model.NewMerchant(l.svcCtx.MyDB)
		if merchant, err = ux.GetMerchantByCode(jwtMerchantCode); err != nil {
			return nil, err
		} else if merchant.AgentLayerCode == "" {
			return nil, errorz.New(response.AGENT_LAYER_NO_GET_ERROR)
		}
		db = db.Where("agent_layer_no LIKE ?", merchant.AgentLayerCode+"%")
		//terms = append(terms, fmt.Sprintf(" agent_layer_no LIKE '%s'", merchant.AgentLayerCode+"%"))
	}
	if len(req.MerchantCode) > 0 {
		db = db.Where("merchant_code = ?", req.MerchantCode)
		//terms = append(terms, fmt.Sprintf(" merchant_code = '%s'", req.MerchantCode))
	}
	if len(req.AgentLayerNo) > 0 {
		db = db.Where("agent_layer_no LIKE ?", "%"+req.AgentLayerNo+"%")
		//terms = append(terms, fmt.Sprintf(" agent_layer_no like '%%%s%%'", req.AgentLayerNo))
	}
	if len(req.Month) > 0 {
		db = db.Where("month = ?", req.Month)
		//terms = append(terms, fmt.Sprintf(" month = '%s'", req.Month))
	}
	if len(req.CurrencyCode) > 0 {
		db = db.Where("currency_code = ?", req.CurrencyCode)
		//terms = append(terms, fmt.Sprintf(" currency_code = '%s'", req.CurrencyCode))
	}

	//term := strings.Join(terms, " AND ")
	//db.Table("cm_commission_month_reports").Where(term)

	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	if err = db.Scopes(gormx.Paginate(*req)).Scopes(gormx.Sort(req.Orders)).Find(&reports).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for _, report := range reports {
		if report.ChangeCommission > 0 {
			totalCommissionAmount += report.ChangeCommission
		} else {
			totalCommissionAmount += report.TotalCommission
		}

	}

	resp = &types.CommissionMonthReportQueryAllResponseX{
		List:                  reports,
		PageNum:               req.PageNum,
		TotalCommissionAmount: totalCommissionAmount,
		PageSize:              req.PageSize,
		RowCount:              count,
	}
	return
}
