package merchant

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantChannelNotifyQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantChannelNotifyQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantChannelNotifyQueryLogic {
	return MerchantChannelNotifyQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantChannelNotifyQueryLogic) MerchantChannelNotifyQuery(req *types.MerchantChannelNotifyQueryRequest) (resp *types.MerchantChannelNotifyQueryRepsonse, err error) {
	var terms []string

	db := l.svcCtx.MyDB.Table("mc_channel_change_notify a").
		Joins("join mc_merchants b on a.merchant_code = b.code")

	if len(req.MerchantCode) > 0 {
		terms = append(terms, fmt.Sprintf("a.merchant_code = '%s'", req.MerchantCode))
		db = db.Where("a.merchant_code = ?", req.MerchantCode)
	}
	if len(req.ParentMerchantCode) > 0 {
		terms = append(terms, fmt.Sprintf("b.agent_parent_code = '%s'", req.ParentMerchantCode))
		db = db.Where("b.agent_parent_code = ?", req.ParentMerchantCode)
	}
	if len(req.Status) > 0 {
		terms = append(terms, fmt.Sprintf("a.status = '%s'", req.Status))
		db = db.Where("a.status = ?", req.Status)
	}
	if len(req.ChannelNotify) > 0 {
		terms = append(terms, fmt.Sprintf("a.is_channel_change_notify = '%s'", req.ChannelNotify))
		db = db.Where("a.is_channel_change_notify = ?", req.ChannelNotify)
	}

	selectX := "a.merchant_code AS merchant_code," +
		"a.is_channel_change_notify AS notify_function," +
		"a.notify_url AS notify_url," +
		"a.last_notify_message AS content," +
		"a.status AS notify_status," +
		"b.agent_parent_code AS parent_merchant_code"

	var merchantChannelNofityQueries []types.MerchantChannelNotifyQuery
	var count int64
	if len(terms) > 0 {
		//term := strings.Join(terms, " AND ")
		if err = db.Count(&count).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		if err = db.Select(selectX).
			Scopes(gormx.Paginate(*req)).
			Find(&merchantChannelNofityQueries).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}

	} else {
		if err = db.Count(&count).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		if err = db.Select(selectX).
			Find(&merchantChannelNofityQueries).Scopes(gormx.Paginate(*req)).Error; err != nil {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	resp = &types.MerchantChannelNotifyQueryRepsonse{
		List:     merchantChannelNofityQueries,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
