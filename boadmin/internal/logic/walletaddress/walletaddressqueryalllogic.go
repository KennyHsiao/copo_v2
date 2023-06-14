package walletaddress

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type WalletAddressQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWalletAddressQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) WalletAddressQueryAllLogic {
	return WalletAddressQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WalletAddressQueryAllLogic) WalletAddressQueryAll(req types.WalletAddressQueryAllRequestX) (resp *types.WalletAddressQueryAllResponseX, err error) {
	var walletAddress []types.WalletAddressX
	var count int64
	//var terms []string
	db := l.svcCtx.MyDB
	db2 := l.svcCtx.MyDB
	if len(req.ChannelPayTypesCode) > 0 {
		//terms = append(terms, fmt.Sprintf("`channel_pay_types_code` like '%%%s%%'", req.ChannelPayTypesCode))
		db = db.Where("channel_pay_types_code like ?", req.ChannelPayTypesCode)
		db2 = db2.Where("channel_pay_types_code like ?", req.ChannelPayTypesCode)
	}
	if len(req.ChannelCode) > 0 {
		//terms = append(terms, fmt.Sprintf("channel_code like '%%%s%%'", req.ChannelCode))
		db = db.Where("channel_code like ?", "%"+req.ChannelCode+"%")
		db2 = db2.Where("channel_code like ?", "%"+req.ChannelCode+"%")
	}
	if len(req.PayTypeCode) > 0 {
		//terms = append(terms, fmt.Sprintf("pay_type_code like '%%%s%%'", req.PayTypeCode))
		db = db.Where("pay_type_code like ?", "%"+req.PayTypeCode+"%")
		db2 = db2.Where("pay_type_code like ?", "%"+req.PayTypeCode+"%")
	}
	if len(req.Account) > 0 {
		//terms = append(terms, fmt.Sprintf("account like '%%%s%%'", req.Account))
		db = db.Where("account like ?", "%"+req.Account+"%")
		db2 = db2.Where("account like ?", "%"+req.Account+"%")
	}
	if len(req.Address) > 0 {
		//terms = append(terms, fmt.Sprintf("address like '%%%s%%'", req.Address))
		db = db.Where("address like ?", "%"+req.Address+"%")
		db2 = db2.Where("address like ?", "%"+req.Address+"%")
	}
	if len(req.Status) > 0 {
		//terms = append(terms, fmt.Sprintf("status = '%s'", req.Status))
		db = db.Where("status = ?", req.Status)
		db2 = db2.Where("status = ?", req.Status)
	}
	if len(req.OrderNo) > 0 {
		//terms = append(terms, fmt.Sprintf("order_no like '%%%s%%'", req.OrderNo))
		db = db.Where("order_no like ?", "%"+req.OrderNo+"%")
		db2 = db2.Where("order_no like ?", "%"+req.OrderNo+"%")
	}

	//term := strings.Join(terms, " AND ")

	if err = db2.Table("ch_wallet_address").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("ch_wallet_address").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		//Order("CONVERT(name USING GBK)").
		Find(&walletAddress).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.WalletAddressQueryAllResponseX{
		List:     walletAddress,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
