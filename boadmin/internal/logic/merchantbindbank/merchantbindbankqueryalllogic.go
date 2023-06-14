package merchantbindbank

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantBindBankQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantBindBankQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantBindBankQueryAllLogic {
	return MerchantBindBankQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantBindBankQueryAllLogic) MerchantBindBankQueryAll(req types.MerchantBindBankQueryAllRequest) (resp *types.MerchantBindBankQueryAllResponse, err error) {
	var merchantBindBanks []types.MerchantBindBank
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("mc_merchant_bind_bank")

	// JWT从登入取得MerchantCode资讯
	merchantCode := l.ctx.Value("merchantCode").(string)
	//terms = append(terms, fmt.Sprintf("merchant_code = '%s'", merchantCode))
	db = db.Where("merchant_code = ?", merchantCode)

	if len(req.CurrencyCode) > 0 {
		//terms = append(terms, fmt.Sprintf("currency_code = '%s'", req.CurrencyCode))
		db = db.Where("currency_code = ?", req.CurrencyCode)
	}
	//term := strings.Join(terms, " AND ")
	db.Count(&count)
	if err = db.Scopes(gormx.Paginate(req)).Find(&merchantBindBanks).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantBindBankQueryAllResponse{
		List:     merchantBindBanks,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
