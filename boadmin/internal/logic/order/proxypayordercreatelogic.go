package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/service/merchantbalanceservice"
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/common/apimodel/vo"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/jinzhu/copier"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/text/language"
	"regexp"

	//orderfeeprofitservice "com.copo/bo_service/boadmin/internal/service/orderfeeprofitservice"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type ProxyPayOrderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProxyPayOrderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ProxyPayOrderCreateLogic {
	return ProxyPayOrderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProxyPayOrderCreateLogic) ProxyPayOrderCreate(req types.MultipleOrderProxyCreateRequest) (resp *types.MultipleOrderCreateResponse, err error) {
	//JWT取得商戶與使用這資訊
	merchantCode := l.ctx.Value("merchantCode").(string)
	userAccount := l.ctx.Value("account").(string)
	var orders = req.List
	var size = len(orders)
	var orderNos []string
	var idxs []string
	var errs []string
	orderNoMap := make(map[string]string)

	for size != len(orderNoMap) {
		//产生单号
		orderNo := model.GenerateOrderNo("DF")
		if _, isExist := orderNoMap[orderNo]; !isExist {
			orderNoMap[orderNo] = orderNo
		}
	}

	for _, v := range orderNoMap {
		orderNos = append(orderNos, v) // 写入交易日志
	}

	// 取得商户检查费率状态
	var merchant types.Merchant
	if err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
		return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
	}

	// 取得商户设定渠道与费率资讯
	merchantOrderRateListViews, err1 := ordersService.GetMerchantChannelRate(l.svcCtx.MyDB, merchantCode, orders[0].CurrencyCode, constants.ORDER_TYPE_DF)
	if err1 != nil {
		return nil, err1
	}
	var merchantOrderRateListView *types.MerchantOrderRateListViewX
	if len(merchantOrderRateListViews) == 1 {
		if req.TypeSubNo == merchantOrderRateListViews[0].DesignationNo {
			merchantOrderRateListView = merchantOrderRateListViews[0]
		} else {
			return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
		}
	} else if len(merchantOrderRateListViews) > 1 {
		if len(req.TypeSubNo) > 0 {
			channelRateMap := make(map[string]*types.MerchantOrderRateListViewX)
			for _, view := range merchantOrderRateListViews {
				channelRateMap[view.DesignationNo] = view
			}
			if _, ok := channelRateMap[req.TypeSubNo]; !ok {
				return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
			} else {
				merchantOrderRateListView = channelRateMap[req.TypeSubNo]
			}
		}
	} else {
		return nil, errorz.New(response.RATE_NOT_CONFIGURED_OR_CHANNEL_NOT_CONFIGURED)
	}

	if merchant.RateCheck != "0" {
		if merchantOrderRateListView.CptHandlingFee > merchantOrderRateListView.MerHandlingFee || merchantOrderRateListView.CptFee > merchantOrderRateListView.MerFee { // 渠道費率與手續費不得高於商戶所設定的
			return nil, errorz.New(response.RATE_SETTING_ERROR)
		}
	}

	//取得是哪種錢包
	balanceType, err2 := merchantbalanceservice.GetBalanceType(l.svcCtx.MyDB, merchantOrderRateListView.ChannelCode, constants.ORDER_TYPE_DF)
	if err2 != nil {
		return nil, err2
	}

	// 取得系统内所有银行
	bankMap := make(map[string]types.ChannelBankX)
	var channelBanks []types.ChannelBankX
	if err = l.svcCtx.MyDB.Table("bk_banks").Select("bank_no, bank_name").Find(&channelBanks).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	for _, bank := range channelBanks {
		bankMap[bank.BankName] = bank
	}

	for i, order := range orders {
		db := l.svcCtx.MyDB

		//验证银行卡号(必填)(必须为数字)(长度必须在10~22码)
		isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, order.MerchantBankAccount)
		currencyCode := order.CurrencyCode
		if currencyCode == constants.CURRENCY_THB {
			if order.MerchantBankAccount == "" || len(order.MerchantBankAccount) < 10 || len(order.MerchantBankAccount) > 16 || !isMatch2 {
				logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需10-16碼內：", order.MerchantBankAccount)
				return nil, errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+order.MerchantBankAccount)
			}
		} else if currencyCode == constants.CURRENCY_CNY {
			if order.MerchantBankAccount == "" || len(order.MerchantBankAccount) < 13 || len(order.MerchantBankAccount) > 22 || !isMatch2 {
				logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需13-22碼內：", order.MerchantBankAccount)
				return nil, errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+order.MerchantBankAccount)
			}
		}

		//判斷黑名單，收款與付款都要判斷
		ux := model.NewBankBlockAccount(db)
		isBlockReceive, err3 := ux.CheckIsBlockAccount(order.MerchantBankAccount)
		if isBlockReceive || err3 != nil {
			logx.WithContext(l.ctx).Errorf("错误:此账号为黑名单，账号: ", order.MerchantBankAccount)
			idxs = append(idxs, order.Index)
			continue
		} else if err3 != nil {
			logx.WithContext(l.ctx).Errorf("錯誤:判断黑名单错误，Index : %v, err : %v", order.Index, err3.Error())
			errs = append(errs, order.Index)
			continue
		}

		// 判断银行名称是否有在系统内
		_, isExist := bankMap[order.MerchantBankName]
		if !isExist {
			logx.WithContext(l.ctx).Errorf("錯誤:判断银行名称系统不存在，Index : %v", order.Index)
			errs = append(errs, order.Index)
			continue
		}

		var errRpc error
		var res *transaction.ProxyOrderUIResponse
		//產生rpc 需要的請求的資料物件
		ProxyOrderUI, rateRpc := l.generateRpcData(order, merchantOrderRateListView, merchantCode, userAccount, orderNos[i])

		if err := transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
			MerchantCode: merchantCode,
			//MerchantOrderNo: req.OrderNo,
			OrderNo:   orderNos[i],
			LogType:   constants.MERCHANT_REQUEST,
			LogSource: constants.PLATEFORM_DF,
			Content:   ProxyOrderUI,
			TraceId:   trace.SpanContextFromContext(l.ctx).TraceID().String(),
		}); err != nil {
			logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", err)
		}

		if balanceType == "DFB" {
			res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderUITransaction_DFB(l.ctx, &transaction.ProxyOrderUIRequest{
				ProxyOrderUI:              ProxyOrderUI,
				MerchantOrderRateListView: rateRpc,
			})
		} else if balanceType == "XFB" {
			res, errRpc = l.svcCtx.TransactionRpc.ProxyOrderUITransaction_XFB(l.ctx, &transaction.ProxyOrderUIRequest{
				ProxyOrderUI:              ProxyOrderUI,
				MerchantOrderRateListView: rateRpc,
			})
		}

		if errRpc != nil {
			logx.WithContext(l.ctx).Error("UI ProxyPayOrder Tranaction rpcResp error:%s", errRpc.Error())

			errs = append(errs, order.Index)
		} else if res.Code != response.API_SUCCESS {
			logx.WithContext(l.ctx).Errorf("UI ProxyPayOrde Tranaction error Code:%s, Message:%s", res.Code, res.Message)
			errs = append(errs, order.Index)
		} else if res.Code == response.API_SUCCESS {
			logx.WithContext(l.ctx).Infof("UI代付交易rpc完成，%s 錢包扣款完成: %#v", balanceType, res.ProxyOrderNo)
		}

		// 串接渠道
		var queryErr error
		var respOrder = &types.OrderX{}
		if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
			logx.WithContext(l.ctx).Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
			return nil, errorz.New(response.FAIL, queryErr.Error())
		}
		merReq, rate := l.generateCallChannelData(respOrder.MerchantBankName, merchantOrderRateListView.ChannelPort)

		// call channel (不論是否有成功打到渠道，都要返回給商戶success，一渠道返回訂單狀態決定此訂單狀態(代處理/處理中))
		var errCHN error
		proxyPayRespVO := &vo.ProxyPayRespVO{}
		proxyPayRespVO, errCHN = ordersService.CallChannel(&l.ctx, &l.svcCtx.Config, merReq, respOrder, rate)

		// 返回給商戶物件
		i18n.SetLang(language.English)
		if errCHN != nil || proxyPayRespVO.Code != "0" {
			logx.WithContext(l.ctx).Errorf("代付提單: %s ，渠道返回錯誤: %s, %#v", respOrder.OrderNo, errCHN.Error(), proxyPayRespVO)
			//将商户钱包加回 (merchantCode, orderNO)，更新狀態為失敗單
			var resRpc *transaction.ProxyPayFailResponse
			if balanceType == "DFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_DFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: respOrder.MerchantCode,
					OrderNo:      respOrder.OrderNo,
				})
			} else if balanceType == "XFB" {
				resRpc, errRpc = l.svcCtx.TransactionRpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
					MerchantCode: respOrder.MerchantCode,
					OrderNo:      respOrder.OrderNo,
				})
			}

			//因在transaction_service 已更新過訂單，重新抓取訂單
			if respOrder, queryErr = model.QueryOrderByOrderNo(l.svcCtx.MyDB, res.ProxyOrderNo, ""); queryErr != nil {
				logx.WithContext(l.ctx).Errorf("撈取代付訂單錯誤: %s, respOrder:%#v", queryErr, respOrder)
				return nil, errorz.New(response.FAIL, queryErr.Error())
			}

			//處理渠道回傳錯誤訊息
			if errCHN != nil {
				respOrder.ErrorType = "2" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
				respOrder.ErrorNote = "渠道异常: " + i18n.Sprintf(errCHN.Error())
				respOrder.Status = constants.FAIL
				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}

			} else if proxyPayRespVO.Code != "0" {
				respOrder.ErrorType = "1" //   1.渠道返回错误	2.渠道异常	3.商户参数错误	4.账户为黑名单	5.其他
				respOrder.ErrorNote = "Code:" + proxyPayRespVO.Code + " Message: " + proxyPayRespVO.Message
				respOrder.Status = constants.FAIL

				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}
			}

			if errRpc != nil {
				logx.WithContext(l.ctx).Errorf("代付提单 %s 还款失败。 Err: %s", respOrder.OrderNo, errRpc.Error())
				respOrder.RepaymentStatus = constants.REPAYMENT_FAIL

				// 更新订单
				if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
					logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
				}

				return nil, errorz.New(response.FAIL, errRpc.Error())
			} else {
				logx.WithContext(l.ctx).Infof("代付還款rpc完成，%s 錢包還款完成: %#v", balanceType, resRpc)
				respOrder.RepaymentStatus = constants.REPAYMENT_SUCCESS
			}

			// 更新订单
			if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
				logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		} else {
			//条整订单状态从"待处理" 到 "交易中"
			respOrder.Status = constants.TRANSACTION
			respOrder.ChannelOrderNo = proxyPayRespVO.Data.ChannelOrderNo

			// 更新订单
			if errUpdate := l.svcCtx.MyDB.Table("tx_orders").Updates(respOrder).Error; errUpdate != nil {
				logx.WithContext(l.ctx).Error("代付订单更新状态错误: ", errUpdate.Error())
			}
		}
	}

	resp = &types.MultipleOrderCreateResponse{
		Index: idxs,
		Errs:  errs,
	}

	return resp, nil
}

func (l *ProxyPayOrderCreateLogic) generateCallChannelData(branchName string, channelPort string) (*types.ProxyPayRequestX, *types.CorrespondMerChnRate) {

	ProxyPayRequest := types.ProxyPayOrderRequest{
		BranchName: branchName,
	}

	ProxyPayRequestX := &types.ProxyPayRequestX{
		ProxyPayOrderRequest: ProxyPayRequest,
	}

	Rate := &types.CorrespondMerChnRate{
		ChannelPort: channelPort,
	}
	return ProxyPayRequestX, Rate
}

func (l *ProxyPayOrderCreateLogic) generateRpcData(req types.OrderProxyCreateRequeset, rate *types.MerchantOrderRateListViewX, merchantCode string, userAccount string, orderNo string) (*transaction.ProxyOrderUI, *transaction.MerchantOrderRateListView) {

	ProxyOrderUI := &transaction.ProxyOrderUI{
		MerchantCode:         merchantCode,
		UserAccount:          userAccount,
		OrderNo:              orderNo,
		OrderAmount:          req.OrderAmount,
		MerchantBankAccount:  req.MerchantBankAccount,
		MerchantBankNo:       req.MerchantBankNo,
		MerchantBankName:     req.MerchantBankName,
		MerchantAccountName:  req.MerchantAccountName,
		MerchantBankProvince: req.MerchantBankProvince,
		MerchantBankCity:     req.MerchantBankCity,
		CurrencyCode:         req.CurrencyCode,
	}

	rateRpc := &transaction.MerchantOrderRateListView{}
	copier.Copy(rateRpc, rate)
	rateRpc.IsRate = rate.CptIsRate

	return ProxyOrderUI, rateRpc
}
