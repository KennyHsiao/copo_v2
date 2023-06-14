package channelBalanceRecord

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
	"time"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelBalanceRecordScheduleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelBalanceRecordScheduleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelBalanceRecordScheduleCreateLogic {
	return ChannelBalanceRecordScheduleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelBalanceRecordScheduleCreateLogic) ChannelBalanceRecordScheduleCreate() error {
	var channelDataList []types.ChannelData

	if err := l.svcCtx.MyDB.Table("ch_channels").Where("status IN ('1','2')").Find(&channelDataList).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	for _, channelData := range channelDataList {
		nowTime := time.Now().Format("2006-01-02 15")
		l.ChannelBalanceRecordCreate(channelData, nowTime)
	}

	return nil
}

func (l *ChannelBalanceRecordScheduleCreateLogic) ChannelBalanceRecordCreate(channel types.ChannelData, time string) {

	record := types.ChannelBalanceRecord{
		Code:         channel.Code,
		CurrencyCode: channel.CurrencyCode,
		Time:         time,
	}

	span := trace.SpanFromContext(l.ctx)
	if !strings.EqualFold(channel.ProxyPayQueryBalanceUrl, "") {
		ProxyKey, errk := utils.MicroServiceEncrypt(l.svcCtx.Config.ApiKey.ProxyKey, l.svcCtx.Config.ApiKey.PublicKey)
		if errk != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 取得系统内部验签错误:%s", channel.Name, errk.Error())
			record.Balance = channel.ProxypayBalance
			record.IsSuccess = "0"
		}
		proxyQueryBalanceRespVO := &vo.ProxyQueryBalanceRespVO{}
		url := fmt.Sprintf("%s:%s/api/proxy-pay-query-balance-internal", l.svcCtx.Config.Server, channel.ChannelPort)
		if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(nil); chnErr != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传错误:%s", channel.Name, chnErr.Error())
			record.Balance = channel.ProxypayBalance
			record.IsSuccess = "0"
		} else if decErr := chnResp.DecodeJSON(proxyQueryBalanceRespVO); decErr != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传值解析错误:%s", channel.Name, decErr.Error())
			record.Balance = channel.ProxypayBalance
			record.IsSuccess = "0"
		} else if proxyQueryBalanceRespVO.Code != "0" {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传错误:%s", channel.Name, proxyQueryBalanceRespVO.Code)
			record.Balance = channel.ProxypayBalance
			record.IsSuccess = "0"
		} else {
			var proxypayBalance float64 = 0
			var errBalance error
			if proxypayBalance, errBalance = strconv.ParseFloat(proxyQueryBalanceRespVO.Data.ProxyPayBalance, 64); errBalance != nil {
				logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传值解析错误:%s", channel.Name, errBalance.Error())
				record.Balance = channel.ProxypayBalance
				record.IsSuccess = "0"
			} else {
				// 正確取得餘額
				record.Balance = proxypayBalance
				record.IsSuccess = "1"
			}
		}
	} else if !strings.EqualFold(channel.PayQueryBalanceUrl, "") {
		PayKey, errk := utils.MicroServiceEncrypt(l.svcCtx.Config.ApiKey.PayKey, l.svcCtx.Config.ApiKey.PublicKey)
		if errk != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 取得系统内部验签错误:%s", channel.Name, errk.Error())
			record.Balance = channel.WithdrawBalance
			record.IsSuccess = "0"
		}

		queryInternalBalanceResp := &vo.ProxyQueryBalanceRespVO{}
		url := fmt.Sprintf("%s:%s/api/pay-query-balance-internal", l.svcCtx.Config.Server, channel.ChannelPort)
		if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", PayKey).JSON(nil); chnErr != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传错误:%s", channel.Name, chnErr.Error())
			record.Balance = channel.WithdrawBalance
			record.IsSuccess = "0"
		} else if decErr := chnResp.DecodeJSON(queryInternalBalanceResp); decErr != nil {
			logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 渠道回传值解析错误:%s", channel.Name, decErr.Error())
			record.Balance = channel.WithdrawBalance
			record.IsSuccess = "0"
		} else {
			var payBalance float64 = 0
			var errBalance error
			if payBalance, errBalance = strconv.ParseFloat(queryInternalBalanceResp.Data.WithdrawBalance, 64); errBalance != nil {
				logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 回传值解析错误:%s", channel.Name, errBalance.Error())
				record.Balance = channel.WithdrawBalance
				record.IsSuccess = "0"
			} else {
				record.Balance = payBalance
				record.IsSuccess = "1"
			}
		}
	} else {
		logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 沒有查詢餘額API", channel.Name)
		return
	}

	if err := l.svcCtx.MyDB.Table("ch_channel_balance_record").Create(&types.ChannelBalanceRecordX{
		ChannelBalanceRecord: record,
	}).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("%s 渠道餘額紀錄 保存失敗:%s", channel.Name, err.Error())
	}
}
