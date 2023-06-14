package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	channelRateRecordService "com.copo/bo_service/boadmin/internal/service"
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"sort"
	"strings"
)

type ChannelPayTypeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeUpdateLogic {
	return ChannelPayTypeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeUpdateLogic) ChannelPayTypeUpdate(req types.ChannelPayTypeUpdateRequest) error {
	err := l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		updatedBy := l.ctx.Value("account").(string)

		channelPayType := &types.ChannelPayTypeUpdate{
			ChannelPayTypeUpdateRequest: req,
		}

		payTypeRate := &types.ChannelPayType{}
		if errPayType := db.Table("ch_channel_pay_types").Where("code = ?", req.Code).Find(payTypeRate).Error; errPayType != nil {
			return errorz.New(response.DATABASE_FAILURE)
		}

		if len(req.FixedAmount) > 0 {
			var FixedAmount = strings.Split(req.FixedAmount, ",")
			sort.Strings(FixedAmount)
			channelPayType.FixedAmount = strings.Join(FixedAmount, ",")
		}

		if req.Fee >= 0 { //包含数字0 也会进入判断
			//merchantChannelRate,err:= model.MerchantChannelRate(l.svcCtx.MyDB).GetMinMerChnFeeByPayTypeCode(req.Code)
			merchantChannelRate, err := model.NewMerchantChannelRate(l.svcCtx.MyDB).GetMinMerChnFeeByPayTypeCode(req.Code)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) && req.Fee > merchantChannelRate.Fee {
				return errorz.New(response.SETTING_CHANNEL_RATE_CHARGE_ERROR, "配置渠道成本费率不可高于商戶配置費率")
			}
		}

		if req.HandlingFee >= 0 {
			merchantChannelRate, err := model.NewMerchantChannelRate(l.svcCtx.MyDB).GetMinMerChnHandlingFeeByPayTypeCode(req.Code)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) && req.HandlingFee > merchantChannelRate.HandlingFee {
				return errorz.New(response.SETTING_CHANNEL_FEE_CHARGE_ERROR, "配置渠道成本手續費不可高于商戶配置手續費")
			}
		}

		if err := db.Table("ch_channel_pay_types").
			Select("Fee", "HandlingFee", "MaxInternalCharge", "DailyTxLimit", "DingleMinCharge", "SingleMinCharge", "SingleMaxCharge", "BillDate", "FixedAmount", "Status", "IsProxy", "IsRate", "Device").
			Where("code = ?", req.Code).
			Updates(channelPayType).Error; err != nil {
			return err
		}

		var errRecord error
		if req.Fee != payTypeRate.Fee {
			errRecord = channelRateRecordService.CreateChannelRateRecord(db, &types.ChannelRateRecordCreateRequest{
				ChannelPayTypeCode: req.Code,
				ModifyType:         "1",
				BeforeRate:         payTypeRate.Fee,
				AfterRate:          req.Fee,
				CreatedBy:          updatedBy,
			})
		}

		if req.HandlingFee != payTypeRate.HandlingFee {
			errRecord = channelRateRecordService.CreateChannelRateRecord(db, &types.ChannelRateRecordCreateRequest{
				ChannelPayTypeCode: req.Code,
				ModifyType:         "2",
				BeforeRate:         payTypeRate.HandlingFee,
				AfterRate:          req.HandlingFee,
				CreatedBy:          updatedBy,
			})

		}
		if errRecord != nil {
			return errRecord
		}

		return nil
	})

	// 发送渠道修改通知
	channelData := &types.ChannelData{}
	if err := l.svcCtx.MyDB.Table("ch_channels").Where("code = ?", req.ChannelCode).Take(channelData).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("渠道变更通知商户名单错误，err : '%v'", err.Error())
	}
	merchantsService.ChannelChangeNotify(l.svcCtx.MyDB, l.ctx, l.svcCtx, channelData.CurrencyCode)

	return err
}
