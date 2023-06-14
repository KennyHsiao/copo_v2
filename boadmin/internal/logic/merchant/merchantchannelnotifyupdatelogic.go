package merchant

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm/clause"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantChannelNotifyUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMerchantChannelNotifyUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantChannelNotifyUpdateLogic {
	return MerchantChannelNotifyUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantChannelNotifyUpdateLogic) MerchantChannelNotifyUpdate(req *types.MerchantChannelNotifyUpdateRequest) (resp *types.MerchantChannelNotifyUpdateResponse, err error) {
	//updatedBy := l.ctx.Value("account").(string)
	updatedBy := "test111"

	var channelChangeNotifyX types.ChannelChangeNotifyX
	channelChangeNotifyX.MerchantCode = req.MerchantCode
	channelChangeNotifyX.IsChannelChangeNotify = req.IsChannelChangeNotify
	channelChangeNotifyX.NotifyUrl = req.NotifyUrl
	channelChangeNotifyX.UpdatedBy = updatedBy

	if err = l.svcCtx.MyDB.Table("mc_channel_change_notify").
		Clauses(clause.Returning{}).
		Select("notify_url", "is_channel_change_notify", "updated_by").
		Where("merchant_code = ?", req.MerchantCode).
		Updates(&channelChangeNotifyX).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	var changeNotifyX types.ChannelChangeNotifyX
	if err = l.svcCtx.MyDB.Table("mc_channel_change_notify").
		Where("merchant_code = ?", req.MerchantCode).Find(&changeNotifyX).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.MerchantChannelNotifyUpdateResponse{
		MerchantCode:          changeNotifyX.MerchantCode,
		IsChannelChangeNotify: changeNotifyX.IsChannelChangeNotify,
		NotifyUrl:             changeNotifyX.NotifyUrl,
		Status:                changeNotifyX.Status,
		LastNotifyMessage:     changeNotifyX.LastNotifyMessage,
	}
	return
}
