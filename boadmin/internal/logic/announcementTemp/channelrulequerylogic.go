package announcementTemp

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelRuleQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelRuleQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelRuleQueryLogic {
	return ChannelRuleQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelRuleQueryLogic) ChannelRuleQuery(req types.ChannelRuleQueryRequestX) (resp *types.ChannelRuleQueryResponseX, err error) {
	var channelRuleQueries []types.ChannelRuleQueryX
	var count int64
	db := l.svcCtx.MyDB

	if req.ID > 0 {
		db = db.Where("id = ?", req.ID)
	}

	if len(req.CostomizeChannelName) > 0 {
		db = db.Where("costomize_channel_name = ?", req.CostomizeChannelName)
	}

	if len(req.ChannelPayType) > 0 {
		db = db.Where("channel_pay_type = ?", req.ChannelPayType)
	}

	if len(req.ChannelMode) > 0 {
		db = db.Where("channel_mode = ?", req.ChannelMode)
	}

	if len(req.Content) > 0 {
		db = db.Where("(chinese_all like ? OR chinese_short like ? OR english_all like ? OR english_short like ?)", req.Content, req.Content, req.Content, req.Content)
	}

	if err = db.Table("an_channel_rule").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("an_channel_rule").
		Scopes(gormx.Paginate(req)).
		Scopes(gormx.Sort(req.Orders)).
		Find(&channelRuleQueries).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.ChannelRuleQueryResponseX{
		List:     channelRuleQueries,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return
}
