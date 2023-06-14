package allocorder

import (
	"com.copo/bo_service/boadmin/internal/model"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AllocRecordCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAllocRecordCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) AllocRecordCreateLogic {
	return AllocRecordCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AllocRecordCreateLogic) AllocRecordCreate(req *types.BKProxyCreateRequeset) (resp string, err error) {

	userAccount := l.ctx.Value("account").(string)
	// 取得系统内所有银行
	bankMap := make(map[string]types.ChannelBankX)
	var channelBanks []types.ChannelBankX
	if err = l.svcCtx.MyDB.Table("bk_banks").Select("bank_no, bank_name").Find(&channelBanks).Error; err != nil {
		return "", errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	for _, bank := range channelBanks {
		bankMap[bank.BankName] = bank
	}

	// 判断银行名称是否有在系统内
	_, isExist := bankMap[req.MerchantBankName]
	if !isExist {
		logx.WithContext(l.ctx).Errorf("錯誤:判断银行名称系统不存在，Index : %v", req.MerchantBankName)
		return "", errorz.New("錯誤:判断银行名称系统不存在: " + req.MerchantBankName)
	}

	orderNo := model.GenerateOrderNo("BK")

	var channelPort string
	channelPayType := &types.ChannelPayType{}
	if err := l.svcCtx.MyDB.Table("ch_channels").Select("channel_port").Where("code = ?", req.ChannelCode).Take(&channelPort).Error; err != nil {
		return "", errorz.New(response.DATA_NOT_FOUND)
	}

	//计算渠道费率与手续费
	if err := l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("code = ?", req.ChannelCode+"DF").Take(channelPayType).Error; err != nil {
		return "", errorz.New(response.DATA_NOT_FOUND)
	}

	//判断单笔最大最小金额
	if req.OrderAmount < channelPayType.SingleMinCharge {
		//金额超过上限
		logx.WithContext(l.ctx).Errorf("錯誤:代付金額未達下限")
		return "", errorz.New(response.ORDER_AMOUNT_LIMIT_MIN)
	} else if req.OrderAmount > channelPayType.SingleMaxCharge {
		//下发金额未达下限
		logx.WithContext(l.ctx).Errorf("錯誤:代付金額超過上限")
		return "", errorz.New(response.ORDER_AMOUNT_LIMIT_MAX)
	}

	var transferHandlingFee float64
	if channelPayType.IsRate == "1" { // 是否算費率，0:否 1:是
		//  交易手續費總額 = 訂單金額 / 100 * 費率 + 手續費
		transferHandlingFee =
			utils.FloatAdd(utils.FloatMul(utils.FloatDiv(req.OrderAmount, 100), channelPayType.Fee), channelPayType.HandlingFee)
	} else {
		//  交易手續費總額 = 訂單金額 / 100 * 費率 + 手續費
		transferHandlingFee =
			utils.FloatAdd(utils.FloatMul(utils.FloatDiv(req.OrderAmount, 100), 0), channelPayType.HandlingFee)
	}

	//建立订单
	txOrder := &types.Order{
		//MerchantCode:         req.MerchantCode,
		CreatedBy:            userAccount,
		MerchantOrderNo:      orderNo,
		OrderNo:              orderNo,
		OrderAmount:          req.OrderAmount,
		BalanceType:          constants.DF_BALANCE,
		Type:                 constants.ORDER_TYPE_BK,
		Status:               constants.WAIT_PROCESS,
		Source:               constants.UI,
		IsMerchantCallback:   constants.MERCHANT_CALL_BACK_DONT_USE,
		IsCalculateProfit:    constants.IS_CALCULATE_PROFIT_NO,
		IsTest:               constants.IS_TEST_NO, //是否測試單
		PersonProcessStatus:  constants.PERSON_PROCESS_STATUS_NO_ROCESSING,
		RepaymentStatus:      constants.REPAYMENT_NOT,
		MerchantBankAccount:  req.MerchantBankAccount,
		MerchantBankNo:       req.MerchantBankNo,
		MerchantBankName:     req.MerchantBankName,
		MerchantAccountName:  req.MerchantAccountName,
		MerchantBankProvince: req.MerchantBankProvince,
		MerchantBankCity:     req.MerchantBankCity,
		CurrencyCode:         req.CurrencyCode,
		ChannelCode:          req.ChannelCode,
		ChannelPayTypesCode:  req.ChannelCode + "DF",
		Fee:                  channelPayType.Fee,
		HandlingFee:          channelPayType.HandlingFee,
		TransferHandlingFee:  transferHandlingFee,
		PayTypeCode:          "DF",
		IsLock:               "0",
	}

	if err = l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		// 创建订单
		if err = db.Table("tx_orders").Create(&types.OrderX{
			Order: *txOrder,
		}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("新增拨款代付UI提单失败, 订单号: %s, err : %s", txOrder.OrderNo, err.Error())
			return
		}

		if err = l.CalculateSystemProfit(db, txOrder, channelPayType, transferHandlingFee); err != nil {
			logx.Errorf("审核通过，计算下发利润失败，商户号: %s, 订单号: %s, err : %s", txOrder.MerchantCode, txOrder.OrderNo, err.Error())
			return err
		}

		return nil
	}); err != nil {
		return "", errorz.New("新增拨款代付UI提单失败")
	}

	order := &types.OrderX{}
	if err := l.svcCtx.MyDB.Table("tx_orders").Where("order_no = ?", orderNo).Take(order).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("捞取订单错误:%s", orderNo)
	}

	// call channel (不論是否有成功打到渠道，都要返回給商戶success，一渠道返回訂單狀態決定此訂單狀態(代處理/處理中))
	var errCHN error
	proxyPayRespVO := &vo.ProxyPayRespVO{}
	proxyPayRespVO, errCHN = ordersService.CallChannel_BK_PROXY(&l.ctx, &l.svcCtx.Config, order, channelPort)
	if errCHN != nil {
		logx.WithContext(l.ctx).Errorf("拨款提單: %s ，渠道返回錯誤: %s, %#v", txOrder.OrderNo, errCHN.Error(), proxyPayRespVO)
		//resRpc, errRpc = rpc.ProxyOrderTransactionFail_XFB(l.ctx, &transaction.ProxyPayFailRequest{
		//失败单
		order.Status = constants.FAIL
		order.TransAt = types.JsonTime{}.New()
		order.ErrorNote = proxyPayRespVO.Code + ":" + proxyPayRespVO.Message

		if err = l.svcCtx.MyDB.Table("tx_orders").Updates(order).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("代付出款失敗(代付餘額)_ %s 更新订单失败: %s", order.OrderNo, err.Error())
			return
		}

		return "", errorz.New(response.CHANNEL_REPLY_ERROR, errCHN.Error())
	} else {
		//条整订单状态从"待处理" 到 "交易中"
		order.Status = constants.TRANSACTION
		order.ChannelOrderNo = proxyPayRespVO.Data.ChannelOrderNo

		if err = l.svcCtx.MyDB.Table("tx_orders").Updates(order).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("代付出款失敗(代付餘額)_ %s 更新订单失败: %s", order.OrderNo, err.Error())
			return
		}
	}

	if err = l.svcCtx.MyDB.Table("tx_orders").Updates(order).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("代付出款失敗(代付餘額)_ %s 更新订单失败: %s", order.OrderNo, err.Error())
		return
	}

	return "success", nil
}

func (l *AllocRecordCreateLogic) CalculateSystemProfit(db *gorm.DB, order *types.Order, channelPayType *types.ChannelPayType, transferHandlingFee float64) (err error) {

	systemFeeProfit := types.OrderFeeProfit{
		OrderNo:             order.OrderNo,
		MerchantCode:        "00000000",
		BalanceType:         order.BalanceType,
		Fee:                 channelPayType.Fee,
		HandlingFee:         channelPayType.HandlingFee,
		TransferHandlingFee: transferHandlingFee,
		// 商戶手續費 - 渠道總手續費 = 利潤 (有可能是負的)
		ProfitAmount: utils.FloatSub(0, transferHandlingFee),
	}

	// 保存系統利潤
	if err = db.Table("tx_orders_fee_profit").Create(&types.OrderFeeProfitX{
		OrderFeeProfit: systemFeeProfit,
	}).Error; err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = l.updateOrderByIsCalculateProfit(db, order.OrderNo); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	return
}

func (l *AllocRecordCreateLogic) updateOrderByIsCalculateProfit(db *gorm.DB, orderNo string) error {
	return db.Table("tx_orders").
		Where("order_no = ?", orderNo).
		Updates(map[string]interface{}{"is_calculate_profit": constants.IS_CALCULATE_PROFIT_YES}).Error
}
