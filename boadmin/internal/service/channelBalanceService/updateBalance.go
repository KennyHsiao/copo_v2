package channelBalanceBalance

import (
	"com.copo/bo_service/boadmin/internal/service/channelDataService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"strings"
	"time"
)

func UpdateChannelBalance(context context.Context, svcCtx *svc.ServiceContext, req *types.ChannelBalanceUpdateRequest) error {

	//channelDataList := &[]types.ChannelData{}

	channelDataList, err := channelDataService.Query_N_ChannelData(svcCtx.MyDB, req.ChannelCodeList)
	if err != nil {
		return err
	}
	span := trace.SpanFromContext(context)
	for _, channel := range *channelDataList {
		//請求代付餘額
		if !strings.EqualFold(channel.ProxyPayQueryBalanceUrl, "") {
			ProxyKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.ProxyKey, svcCtx.Config.ApiKey.PublicKey)
			if errk != nil {
				return errorz.New(response.GENERAL_EXCEPTION, errk.Error())
			}
			proxyQueryBalanceRespVO := &vo.ProxyQueryBalanceRespVO{}
			url := fmt.Sprintf("%s:%s/api/proxy-pay-query-balance-internal", svcCtx.Config.Server, channel.ChannelPort)
			if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(nil); chnErr != nil {
				return chnErr
			} else if decErr := chnResp.DecodeJSON(proxyQueryBalanceRespVO); decErr != nil {
				return decErr
			} else if proxyQueryBalanceRespVO.Code != "0" {
				return errorz.New(response.UPDATE_CHANNEL_BALANCE_ERROR, proxyQueryBalanceRespVO.Data.ChannelCodingtring)
			}

			var proxypayBalance float64 = 0
			var errBalance error
			if proxypayBalance, errBalance = strconv.ParseFloat(proxyQueryBalanceRespVO.Data.ProxyPayBalance, 64); errBalance != nil {
				return errBalance
			}

			channel.ProxypayBalance = proxypayBalance
		}

		if !strings.EqualFold(channel.PayQueryBalanceUrl, "") {
			PayKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.PayKey, svcCtx.Config.ApiKey.PublicKey)
			if errk != nil {
				return errorz.New(response.GENERAL_EXCEPTION, errk.Error())
			}

			queryInternalBalanceResp := &vo.ProxyQueryBalanceRespVO{}
			url := fmt.Sprintf("%s:%s/api/pay-query-balance-internal", svcCtx.Config.Server, channel.ChannelPort)
			if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", PayKey).JSON(nil); chnErr != nil {
				return chnErr
			} else if decErr := chnResp.DecodeJSON(queryInternalBalanceResp); decErr != nil {
				return decErr
			}

			var payBalance float64 = 0
			var errBalance error
			if payBalance, errBalance = strconv.ParseFloat(queryInternalBalanceResp.Data.WithdrawBalance, 64); errBalance != nil {
				return errBalance
			}
			channel.WithdrawBalance = payBalance
		}
		channel.UpdatedAt = time.Now().UTC()
		if errSave := svcCtx.MyDB.Table("ch_channels").Save(&channel); err != nil {
			return errSave.Error
		}

	}

	return nil
}

func UpdateChannelBalanceForSchedule(context context.Context, svcCtx *svc.ServiceContext, channelDataList []types.ChannelDataUpdate) (resp types.ChannelBalanceNotify, err error) {
	var successList []string
	var failList []string
	span := trace.SpanFromContext(context)
	for _, channel := range channelDataList {
		//請求代付餘額
		if !strings.EqualFold(channel.ProxyPayQueryBalanceUrl, "") {
			ProxyKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.ProxyKey, svcCtx.Config.ApiKey.PublicKey)
			if errk != nil {
				logx.WithContext(context).Errorf("更新代付馀额取得系统内部验签错误:%s", errk.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}
			proxyQueryBalanceRespVO := &vo.ProxyQueryBalanceRespVO{}
			url := fmt.Sprintf("%s:%s/api/proxy-pay-query-balance-internal", svcCtx.Config.Server, channel.ChannelPort)
			if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", ProxyKey).JSON(nil); chnErr != nil {
				logx.WithContext(context).Errorf("更新代付馀额渠道回传错误:%s", chnErr.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			} else if decErr := chnResp.DecodeJSON(proxyQueryBalanceRespVO); decErr != nil {
				logx.WithContext(context).Errorf("更新代付馀额渠道回传值解析错误:%s", decErr.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			} else if proxyQueryBalanceRespVO.Code != "0" {
				logx.WithContext(context).Errorf("更新代付馀额渠道回传错误:%s", proxyQueryBalanceRespVO.Code)
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}

			var proxypayBalance float64 = 0
			var errBalance error
			if proxypayBalance, errBalance = strconv.ParseFloat(proxyQueryBalanceRespVO.Data.ProxyPayBalance, 64); errBalance != nil {
				logx.WithContext(context).Errorf("更新代付馀额渠道回传值解析错误:%s", errBalance.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}

			channel.ProxypayBalance = proxypayBalance
		}

		if !strings.EqualFold(channel.PayQueryBalanceUrl, "") {
			PayKey, errk := utils.MicroServiceEncrypt(svcCtx.Config.ApiKey.PayKey, svcCtx.Config.ApiKey.PublicKey)
			if errk != nil {
				logx.WithContext(context).Errorf("更新支付馀额取得系统内部验签错误:%s", errk.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}

			queryInternalBalanceResp := &vo.ProxyQueryBalanceRespVO{}
			url := fmt.Sprintf("%s:%s/api/pay-query-balance-internal", svcCtx.Config.Server, channel.ChannelPort)
			if chnResp, chnErr := gozzle.Post(url).Timeout(10).Trace(span).Header("authenticationProxyPaykey", PayKey).JSON(nil); chnErr != nil {
				logx.WithContext(context).Errorf("更新支付馀额渠道回传错误:%s", chnErr.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			} else if decErr := chnResp.DecodeJSON(queryInternalBalanceResp); decErr != nil {
				logx.WithContext(context).Errorf("更新支付馀额渠道回传值解析错误:%s", decErr.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}

			var payBalance float64 = 0
			var errBalance error
			if payBalance, errBalance = strconv.ParseFloat(queryInternalBalanceResp.Data.WithdrawBalance, 64); errBalance != nil {
				logx.WithContext(context).Errorf("更新支付馀额渠道回传值解析错误:%s", errBalance.Error())
				failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
				continue
			}
			channel.WithdrawBalance = payBalance
		}
		channel.UpdatedAt = time.Now().UTC()
		if errSave := svcCtx.MyDB.Table("ch_channels").Save(&channel); errSave.Error != nil {
			logx.WithContext(context).Errorf("更新馀额渠道DB错误:%s", errSave.Error)
			failList = append(failList, fmt.Sprintf(channel.Name+"("+channel.CurrencyCode+")"))
			continue
		}
		successList = append(successList, channel.Code)
	}

	resp = types.ChannelBalanceNotify{
		SuccessList: successList,
		FailList:    failList,
	}

	return resp, nil
}
