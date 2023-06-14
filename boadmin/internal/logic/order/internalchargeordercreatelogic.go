package order

import (
	"com.copo/bo_service/boadmin/internal/model"
	orderfeeprofitservice "com.copo/bo_service/boadmin/internal/service/orderfeeprofitservice"
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"context"
	"github.com/copo888/transaction_service/rpc/transaction"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

type InternalChargeOrderCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInternalChargeOrderCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) InternalChargeOrderCreateLogic {
	return InternalChargeOrderCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InternalChargeOrderCreateLogic) InternalChargeOrderCreate(req types.OrderInternalCreate) error {
	imgUrl, err := MultifileUpload(req)
	imgUrl = strings.Replace(imgUrl, "./public/", "", -1)
	logx.Info("圖片位置:", imgUrl)

	if err != nil {
		return errorz.New(response.FILE_UPLOAD_ERROR, err.Error())
	}

	// JWT取得登入腳色資訊 用於商戶號與createdBy
	merchantCode := l.ctx.Value("merchantCode").(string)
	userAccount := l.ctx.Value("account").(string)
	//判斷黑名單，收款與付款都要判斷
	ux := model.NewBankBlockAccount(l.svcCtx.MyDB)
	isBlockReceive, err1 := ux.CheckIsBlockAccount(req.MerchantBankAccount)
	if err1 != nil {
		RemoveFile(imgUrl)
		return errorz.New(response.DATABASE_FAILURE, err1.Error())
	}

	isBlockPay, err2 := ux.CheckIsBlockAccount(req.ChannelBankAccount)
	if err2 != nil {
		return errorz.New(response.DATABASE_FAILURE, err2.Error())
	}
	if isBlockReceive || isBlockPay {
		RemoveFile(imgUrl)
		return errorz.New(response.BANK_ACCOUNT_IN_BLACK_LIST)
	}

	//验证银行卡号，收款與付款都要判斷(必填)(必须为数字)(长度必须在10~22码)
	isMatch, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.MerchantBankAccount)
	currencyCode := req.CurrencyCode
	if currencyCode == constants.CURRENCY_THB {
		if req.MerchantBankAccount == "" || len(req.MerchantBankAccount) < 10 || len(req.MerchantBankAccount) > 16 || !isMatch {
			logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需10-16碼內：", req.MerchantBankAccount)
			return errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+req.MerchantBankAccount)
		}
	} else if currencyCode == constants.CURRENCY_CNY {
		if req.MerchantBankAccount == "" || len(req.MerchantBankAccount) < 13 || len(req.MerchantBankAccount) > 22 || !isMatch {
			logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需13-22碼內：", req.MerchantBankAccount)
			return errorz.New(response.INVALID_BANK_NO, "MerchantBankAccount: "+req.MerchantBankAccount)
		}
	}
	isMatch2, _ := regexp.MatchString(constants.REGEXP_BANK_ID, req.ChannelBankAccount)
	if currencyCode == constants.CURRENCY_THB {
		if req.ChannelBankAccount == "" || len(req.ChannelBankAccount) < 10 || len(req.ChannelBankAccount) > 16 || !isMatch2 {
			logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需10-16碼內：", req.ChannelBankAccount)
			return errorz.New(response.INVALID_BANK_NO, "ChannelBankAccount: "+req.ChannelBankAccount)
		}
	} else if currencyCode == constants.CURRENCY_CNY {
		if req.ChannelBankAccount == "" || len(req.ChannelBankAccount) < 13 || len(req.ChannelBankAccount) > 22 || !isMatch2 {
			logx.WithContext(l.ctx).Error("銀行卡號檢查錯誤，需13-22碼內： ", req.ChannelBankAccount)
			return errorz.New(response.INVALID_BANK_NO, "ChannelBankAccount: "+req.ChannelBankAccount)
		}
	}

	//// 取得商户检查费率状态
	//var merchant types.Merchant
	//if err := l.svcCtx.MyDB.Table("mc_merchants").Where("code = ?", merchantCode).Take(&merchant).Error; err != nil {
	//	RemoveFile(imgUrl)
	//	return errorz.New(response.DATA_NOT_FOUND, err.Error())
	//}
	//
	//// 取得商户设定渠道与费率资讯
	//merchantOrderRateListView, err3 := ordersService.GetMerchantChannelRate(l.svcCtx.MyDB, merchantCode, req.CurrencyCode, constants.ORDER_TYPE_NC)
	//if err3 != nil {
	//	RemoveFile(imgUrl)
	//	return err3
	//}
	//
	//if merchant.RateCheck != "0" {
	//	if merchantOrderRateListView[0].CptHandlingFee > merchantOrderRateListView[0].MerHandlingFee || merchantOrderRateListView[0].CptFee > merchantOrderRateListView[0].MerFee { // 渠道費率與手續費不得高於商戶所設定的
	//		RemoveFile(imgUrl)
	//		return errorz.New(response.RATE_SETTING_ERROR)
	//	}
	//}

	//// 检查渠道最大内冲金额
	//if merchantOrderRateListView[0].MaxInternalCharge < req.OrderAmount {
	//	RemoveFile(imgUrl)
	//	return errorz.New(response.CHARGE_AMT_EXCEED)
	//}

	var orderReq types.OrderX
	orderReq.InternalChargeOrderPath = imgUrl
	orderReq.OrderAmount = req.OrderAmount
	orderReq.CurrencyCode = req.CurrencyCode
	orderReq.MerchantAccountName = req.MerchantAccountName
	orderReq.MerchantBankAccount = req.MerchantBankAccount
	orderReq.MerchantBankCity = req.MerchantBankCity
	orderReq.MerchantBankProvince = req.MerchantBankProvince
	orderReq.MerchantBankNo = req.MerchantBankNo
	orderReq.MerchantBankName = req.MerchantBankName
	orderReq.ChannelBankNo = req.ChannelBankNo
	orderReq.ChannelBankName = req.ChannelBankName
	orderReq.ChannelAccountName = req.ChannelAccountName
	orderReq.ChannelBankAccount = req.ChannelBankAccount

	//產生rpc 需要的請求的資料物件
	InternalOrder, rateRpc := l.generateRpcData(&orderReq, nil, merchantCode, userAccount)

	var errRpc error
	var res *transaction.InternalOrderResponse

	res, errRpc = l.svcCtx.TransactionRpc.InternalOrderTransaction(l.ctx, &transaction.InternalOrderRequest{
		InternalOrder:             InternalOrder,
		MerchantOrderRateListView: rateRpc,
	})

	if errRpc != nil {
		logx.WithContext(l.ctx).Error("InternalChargeOrder Tranaction rpcResp error:%s", errRpc.Error())
		return errorz.New(response.FAIL, errRpc.Error())
	} else if res.Code != response.API_SUCCESS {
		logx.WithContext(l.ctx).Errorf("InternalChargeOrder Tranaction error Code:%s, Message:%s", res.Code, res.Message)
		return errorz.New(res.Code, res.Message)
	} else if res.Code == response.API_SUCCESS {
		logx.WithContext(l.ctx).Infof("内充交易rpc完成，单号:  %s ", res.OrderNo)
		// 写入交易日志
		if err := transactionLogService.CreateTransactionLog(l.svcCtx.MyDB, &types.TransactionLogData{
			MerchantCode: merchantCode,
			//MerchantOrderNo: "",
			OrderNo: res.OrderNo,
			//ChannelOrderNo:  "",
			LogType:       constants.MERCHANT_REQUEST,
			LogSource:     constants.PLATEFORM_NC,
			TxOrderSource: constants.UI,
			TxOrderType:   constants.ORDER_TYPE_NC,
			Content:       req.OrderX,
			TraceId:       trace.SpanContextFromContext(l.ctx).TraceID().String(),
		}); err != nil {
			logx.WithContext(l.ctx).Errorf("写入交易日志错误:%s", err)
		}
		return nil
	}

	return nil
}

func MultifileUpload(req types.OrderInternalCreate) (resp string, err error) {
	files := req.FormData["uploadFile"]
	var splitStrs []string
	for _, ff := range files {
		ext := strings.ToLower(path.Ext(ff.Filename))
		if ext != ".jpg" && ext != ".png" {
			return "", errorz.New(response.FILE_TYPE_NOT_JPG_ERROR)
		}
		file, err := ff.Open()
		if err != nil {

		}
		defer file.Close()
		var terms []string
		randStr := random.GetRandomString(10, random.ALL, random.MIX)
		terms = append(append(terms, randStr), ext)
		newFileName := strings.Join(terms, "")
		f, errOpenFile := os.OpenFile("./public/uploads/internalcharges/"+newFileName, os.O_WRONLY|os.O_CREATE, 0777)
		if errOpenFile != nil {
			return "nil", errorz.New(response.FAIL, err.Error())
		}
		defer f.Close()
		////把.去掉
		//splitStr := strings.Split(f.Name(), ".")
		io.Copy(f, file)
		splitStrs = append(splitStrs, f.Name())
	}
	term := strings.Join(splitStrs, ",")
	return term, nil
}

func RemoveFile(path string) error {
	splitStr := strings.Split(path, ",")
	for _, s := range splitStr {
		if err := os.Remove(s); err != nil {
			return errorz.New(response.FAIL, err.Error())
		}
	}

	return nil
}

func createInternalOrder(db *gorm.DB, req types.OrderInternalCreate, merchantOrderRateListView types.MerchantOrderRateListViewX,
	merchantCode string, userAccount string, imgUrl string) (err error) {
	orderReq := req.OrderX
	//产生单号
	orderNo := model.GenerateOrderNo("NC")

	// 計算利潤
	var orderFeeProfits []types.OrderFeeProfit
	if orderFeeProfits, err = orderfeeprofitservice.CalculateOrderProfit(db, types.CalculateProfit{
		MerchantCode:        merchantCode,
		OrderNo:             orderNo,
		Type:                constants.ORDER_TYPE_NC,
		CurrencyCode:        req.CurrencyCode,
		BalanceType:         "DFB",
		ChannelCode:         merchantOrderRateListView.ChannelCode,
		ChannelPayTypesCode: merchantOrderRateListView.ChannelPayTypesCode,
		OrderAmount:         req.OrderAmount,
	}); err != nil {
		return err
	}

	//新增订单
	orderReq.OrderNo = orderNo
	orderReq.Type = constants.ORDER_TYPE_NC
	orderReq.MerchantCode = merchantCode
	orderReq.Status = constants.PROCESSING // 0:待處理 1:處理中 20:成功 30:失敗 31:凍結
	orderReq.Source = constants.UI
	//orderReq.CallBackStatus = "1" // 回调状态
	orderReq.InternalChargeOrderPath = imgUrl
	orderReq.BalanceType = "DFB"
	orderReq.OrderAmount = req.OrderAmount
	orderReq.TransferAmount = orderReq.OrderAmount + orderFeeProfits[0].TransferHandlingFee
	orderReq.CreatedBy = userAccount
	orderReq.UpdatedBy = userAccount
	orderReq.IsLock = "0" //是否锁定状态 (0=否;1=是) 预设否
	orderReq.TransferHandlingFee = orderFeeProfits[0].TransferHandlingFee
	orderReq.CurrencyCode = req.CurrencyCode
	orderReq.MerchantAccountName = req.MerchantAccountName
	orderReq.MerchantBankAccount = req.MerchantBankAccount
	orderReq.MerchantBankCity = req.MerchantBankCity
	orderReq.MerchantBankProvince = req.MerchantBankProvince
	orderReq.MerchantBankNo = req.MerchantBankNo
	orderReq.MerchantBankName = req.MerchantBankName
	orderReq.ChannelCode = merchantOrderRateListView.ChannelCode
	orderReq.ChannelBankName = req.ChannelBankName
	orderReq.ChannelAccountName = req.ChannelAccountName
	orderReq.ChannelBankAccount = req.ChannelBankAccount
	orderReq.ChannelBankNo = req.ChannelBankNo
	orderReq.ChannelPayTypesCode = merchantOrderRateListView.ChannelPayTypesCode
	orderReq.PayTypeCode = merchantOrderRateListView.PayTypeCode

	if err := db.Table("tx_orders").Create(&orderReq).Error; err != nil {
		RemoveFile(imgUrl)
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//记录订单历程
	orderAction := types.OrderAction{
		OrderNo:     orderNo,
		Action:      "PLACE_ORDER",
		UserAccount: userAccount,
		Comment:     "",
	}
	if err := model.NewOrderAction(db).CreateOrderAction(&types.OrderActionX{
		OrderAction: orderAction,
	}); err != nil {
		return err
	}
	return nil
}

func (l *InternalChargeOrderCreateLogic) generateRpcData(req *types.OrderX, rate *types.MerchantOrderRateListViewX, merchantCode string, userAccount string) (*transaction.InternalOrder, *transaction.MerchantOrderRateListView) {

	InternalOrder := &transaction.InternalOrder{
		Imgurl:               req.InternalChargeOrderPath,
		MerchantCode:         merchantCode,
		UserAccount:          userAccount,
		OrderAmount:          req.OrderAmount,
		CurrencyCode:         req.CurrencyCode,
		MerchantAccountName:  req.MerchantAccountName,
		MerchantBankAccount:  req.MerchantBankAccount,
		MerchantBankCity:     req.MerchantBankCity,
		MerchantBankProvince: req.MerchantBankProvince,
		MerchantBankNo:       req.MerchantBankNo,
		MerchantBankName:     req.MerchantBankName,
		ChannelBankNo:        req.ChannelBankNo,
		ChannelBankName:      req.ChannelBankName,
		ChannelBankAccount:   req.ChannelBankAccount,
		ChannelAccountName:   req.ChannelAccountName,
	}

	if rate != nil {
		rateRpc := &transaction.MerchantOrderRateListView{
			ChannelPayTypesCode: rate.ChannelPayTypesCode,
			PayTypeCode:         rate.PayTypeCode,
			MerHandlingFee:      rate.MerHandlingFee,
			MerFee:              rate.MerFee,
			Designation:         rate.Designation,
			DesignationNo:       rate.DesignationNo,
			ChannelCode:         rate.ChannelCode,
			CurrencyCode:        rate.CurrencyCode,
			MaxInternalCharge:   rate.MaxInternalCharge,
			SingleMinCharge:     rate.SingleMinCharge,
			SingleMaxCharge:     rate.SingleMaxCharge,
			MerchantCode:        rate.MerchantCode,
			MerchnrateStatus:    rate.MerchnrateStatus,
			ChnStatus:           rate.ChnStatus,
			ChnIsProxy:          rate.ChnIsProxy,
			CptStatus:           rate.CptStatus,
			CptFee:              rate.CptFee,
			CptHandlingFee:      rate.CptHandlingFee,
		}
		return InternalOrder, rateRpc
	} else {
		return InternalOrder, nil
	}
}
