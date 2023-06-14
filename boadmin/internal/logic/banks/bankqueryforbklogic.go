package banks

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type BankQueryForBKLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBankQueryForBKLogic(ctx context.Context, svcCtx *svc.ServiceContext) BankQueryForBKLogic {
	return BankQueryForBKLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BankQueryForBKLogic) BankQueryForBK(req *types.BankQueryForBKRequestX) (resp *types.BankQueryForBKResponse, err error) {

	var banks []types.Bank
	var currencyCode string
	if err := l.svcCtx.MyDB.Table("ch_channels cc").
		Select("cc.currency_code").
		Where("cc.code = ?", req.ChannelCode).
		Take(&currencyCode).Error; err != nil {
		return nil, errorz.New(response.INVALID_CHANNEL_INFO)
	}

	selectX := "bb.id," +
		"bb.bank_no," +
		"bb.bank_name," +
		"bb.bank_name_en," +
		"bb.abbr," +
		"bb.branch_no," +
		"bb.branch_name," +
		"bb.city," +
		"bb.province," +
		"bb.currency_code," +
		"bb.status"

	if err := l.svcCtx.MyDB.Table("bk_banks bb").
		Joins("RIGHT JOIN ch_channel_banks ccb ON ccb.bank_no = bb.bank_no").
		Scopes(gormx.Sort(req.Orders)).
		Select(selectX).
		Where("ccb.channel_code = ?", req.ChannelCode).
		Where("bb.currency_code = ?", currencyCode).
		Find(&banks).Error; err != nil {
		return nil, errorz.New(response.BANK_CODE_INVALID)
	}

	if len(banks) == 0 {

		var currencyCode string

		if err := l.svcCtx.MyDB.Table("ch_channels cc").
			Select("currency_code").
			Where("code = ?", req.ChannelCode).
			Take(&currencyCode).Error; err != nil {

		}

		if err := l.svcCtx.MyDB.Table("bk_banks bb").
			Scopes(gormx.Sort(req.Orders)).
			Where("currency_code = ?", currencyCode).
			Find(&banks).Error; err != nil {
			return nil, errorz.New(response.BANK_CODE_INVALID)
		}
	}

	return &types.BankQueryForBKResponse{
		List: banks,
	}, nil
}
