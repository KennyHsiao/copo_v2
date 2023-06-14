package commissionWithdrawOrder

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommissionWithdrawOrderQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommissionWithdrawOrderQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) CommissionWithdrawOrderQueryAllLogic {
	return CommissionWithdrawOrderQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommissionWithdrawOrderQueryAllLogic) CommissionWithdrawOrderQueryAll(req types.CommissionWithdrawOrderQueryAllRequestX) (resp *types.CommissionWithdrawOrderQueryAllResponseX, err error) {
	var commissionWithdrawOrders []types.CommissionWithdrawOrderX
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB.Table("cm_withdraw_order")
	resp = &types.CommissionWithdrawOrderQueryAllResponseX{}

	jwtMerchantCode := l.ctx.Value("merchantCode").(string)

	if len(jwtMerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" merchant_code = '%s'", jwtMerchantCode))
		db = db.Where("merchant_code = ?", jwtMerchantCode)
	}
	if len(req.MerchantCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" merchant_code like '%%%s%%'", req.MerchantCode))
		db = db.Where("merchant_code LIKE ?", "%"+req.MerchantCode+"%")
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf(" order_no = '%s'", req.OrderNo))
		db = db.Where("order_no = ?", req.OrderNo)
	}
	if len(req.WithdrawCurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" withdraw_currency_code = '%s'", req.WithdrawCurrencyCode))
		db = db.Where("withdraw_currency_code = ?", req.WithdrawCurrencyCode)
	}
	if len(req.PayCurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf(" pay_currency_code = '%s'", req.PayCurrencyCode))
		db = db.Where("pay_currency_code = ?", req.PayCurrencyCode)
	}
	if len(req.StartAt) > 0 {
		//terms = append(terms, fmt.Sprintf(" created_at >= '%s'", req.StartAt))
		db = db.Where("created_at > ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf(" created_at < '%s'", endAt))
		db = db.Where("created_at < ?", endAt)
	}

	//term := strings.Join(terms, " AND ")

	if err = db.Select(" sum(withdraw_amount) as total_withdraw_amount, " +
		" sum(pay_amount) as total_pay_amount").
		Table("cm_withdraw_order").Find(&resp).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Select("*").Table("cm_withdraw_order").
		Scopes(gormx.Paginate(req)).Scopes(gormx.Sort(req.Orders)).Find(&commissionWithdrawOrders).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for i, order := range commissionWithdrawOrders {
		if order.AttachmentPath != "" {
			commissionWithdrawOrders[i].AttachmentPath = l.svcCtx.Config.ResourceHost + order.AttachmentPath
		}
	}

	resp.List = commissionWithdrawOrders
	resp.PageNum = req.PageNum
	resp.PageSize = req.PageSize
	resp.RowCount = count

	return
}
