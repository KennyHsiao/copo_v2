package etl

import (
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"encoding/json"
	"github.com/gioco-play/go-driver/logrusz"
	"github.com/gioco-play/go-driver/mysqlz"
	"gorm.io/gorm"
	"strconv"

	"com.copo/bo_service/boadmin/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChannelEtlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type ChannelPayMethod struct {
	ChannelEnableStatus    string `json:"channel_enable_status"`
	ChannelRate            string `json:"channel_rate"`
	ChannelHandingFee      string `json:"channel_handing_fee"`
	OneDayTransactionLimit string `json:"one_day_transaction_limit"`
	SingleLimitMinimum     string `json:"single_limit_minimum"`
	SingleLimitMaximum     string `json:"single_limit_maximum"`
	MaximunAmountOfCharge  string `json:"maximun_amount_of_charge"`
	WalletShareType        string `json:"wallet_share_type"`
	FixedAmount            string `json:"fixed_amount"`
	Device                 string `json:"device"`
	PayTypeCoding          string `json:"pay_type_coding"`
	ChannelCoding          string `json:"channel_coding"`
	PayType                string `json:"pay_type"`
	TradingDelayDay        string `json:"trading_delay_day"`
}

type ChannelInfo struct {
	ChannelCoding string `json:"channel_coding"`
	PayTypeMap    string `json:"pay_type_map"`
}

type PayType struct {
	ParamNumber string `json:"param_number"`
	ParamName   string `json:"param_name"`
}

type ChannelData struct {
	ChannelCoding string `json:"channel_coding"`
	IsWalletShare string `json:"is_wallet_share"`
}

type ChannelBalance struct {
	ChannelCoding   string `json:"channel_coding"`
	WithdrawBalance string `json:"withdraw_balance"`
	ProxypayBalance string `json:"proxypay_balance"`
}

type BankMain struct {
	MainBankNo     string `json:"main_bank_no"`
	CurrencyCoding string `json:"currency_coding"`
	MainBankName   string `json:"main_bank_name"`
	EngName        string `json:"eng_name"`
	Abbreviation   string `json:"abbreviation"`
}

type BankChannel struct {
	MainBankNo    string `json:"main_bank_no"`
	ChannelCoding string `json:"channel_coding"`
	BankCode      string `json:"bank_code"`
}

func NewChannelEtlLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelEtlLogic {
	return ChannelEtlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelEtlLogic) ChannelEtl() error {
	oldDb, err := mysqlz.New("8.129.209.41", "3306", "dior", "P#tjnnPEZ@JwQjkFrcdG", "dior22").
		SetCharset("utf8mb4").
		SetLoc("UTC").
		SetLogger(logrusz.New().SetLevel("debug").Writer()).
		Connect(mysqlz.Pool(1, 1, 1))
	if err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	//channelPayMethod导入
	//if err := l.channelPayMethod(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "channelPayMethod导入錯誤："+err.Error())
	//}
	// payType导入
	//if err := l.payType(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "payType导入错误 : "+err.Error())
	//}
	// channel更新是否支轉代
	//if err := l.updateChannelIsWalletShare(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "channel更新是否支轉代 : "+err.Error())
	//}
	// 基础银行资料
	//if err := l.Banks(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "")
	//}
	// 渠道银行对应资料导入
	if err := l.channelBank(oldDb); err != nil {
		return errorz.New(response.DATABASE_FAILURE, err.Error())
	}
	return nil
}

func (l *ChannelEtlLogic) channelPayMethod(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var channelPayMethods []ChannelPayMethod
		if err := oldDb.Table("channel_pay_method").Find(&channelPayMethods).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var channelInfos []ChannelInfo
		if err := oldDb.Table("channel_info").Find(&channelInfos).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var channels []types.ChannelData
		if err := tx.Table("ch_channels").Find(&channels).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		channelInfoMap := make(map[string]ChannelInfo)
		for _, info := range channelInfos {
			channelInfoMap[info.ChannelCoding] = info
		}

		channelMap := make(map[string]string)
		for _, channel := range channels {
			channelMap[channel.Code] = channel.Code
		}

		var channelPayTypeCreates []types.ChannelPayTypeCreate
		for _, method := range channelPayMethods {
			mChannelRate := "0"
			if len(method.ChannelRate) > 0 {
				mChannelRate = method.ChannelRate
			}
			mChannelHandingFee := "0"
			if len(method.ChannelHandingFee) > 0 {
				mChannelHandingFee = method.ChannelHandingFee
			}
			mOneDayTransactionLimit := "0"
			if len(method.OneDayTransactionLimit) > 0 {
				mOneDayTransactionLimit = method.OneDayTransactionLimit
			}
			mSingleLimitMinimum := "0"
			if len(method.SingleLimitMinimum) > 0 {
				mSingleLimitMinimum = method.SingleLimitMinimum
			}
			mSingleLimitMaximum := "0"
			if len(method.SingleLimitMaximum) > 0 {
				mSingleLimitMaximum = method.SingleLimitMaximum
			}
			mMaximunAmountOfCharge := "0"
			if len(method.MaximunAmountOfCharge) > 0 {
				mMaximunAmountOfCharge = method.MaximunAmountOfCharge
			}

			if _, ok := channelMap[method.ChannelCoding]; ok {
				if v, ok2 := channelInfoMap[method.ChannelCoding]; ok2 {
					var payTypeMap map[string]string
					if err := json.Unmarshal([]byte(v.PayTypeMap), &payTypeMap); err != nil {
						return errorz.New(response.SYSTEM_ERROR, err.Error())
					}

					var channelPayTypeCreate types.ChannelPayTypeCreate
					channelPayTypeCreate.Fee, _ = strconv.ParseFloat(mChannelRate, 64)
					channelPayTypeCreate.HandlingFee, _ = strconv.ParseFloat(mChannelHandingFee, 64)
					channelPayTypeCreate.DailyTxLimit, _ = strconv.ParseFloat(mOneDayTransactionLimit, 64)
					channelPayTypeCreate.SingleMinCharge, _ = strconv.ParseFloat(mSingleLimitMinimum, 64)
					channelPayTypeCreate.SingleMaxCharge, _ = strconv.ParseFloat(mSingleLimitMaximum, 64)
					channelPayTypeCreate.MaxInternalCharge, _ = strconv.ParseFloat(mMaximunAmountOfCharge, 64)
					channelPayTypeCreate.IsProxy = method.WalletShareType
					channelPayTypeCreate.FixedAmount = method.FixedAmount
					channelPayTypeCreate.Device = method.Device
					channelPayTypeCreate.Status = method.ChannelEnableStatus
					channelPayTypeCreate.ChannelCode = method.ChannelCoding
					channelPayTypeCreate.Code = method.PayTypeCoding
					channelPayTypeCreate.PayTypeCode = method.PayType
					channelPayTypeCreate.BillDate, _ = strconv.ParseInt(method.TradingDelayDay, 1, 32)
					channelPayTypeCreate.MapCode = ""
					if p, ok3 := payTypeMap[channelPayTypeCreate.PayTypeCode]; ok3 {
						channelPayTypeCreate.MapCode = p
					}

					channelPayTypeCreates = append(channelPayTypeCreates, channelPayTypeCreate)
				}
			}

			//for _, payType := range channelPayTypes {
			//	if method.PayTypeCoding == payType.Code {
			//		payType.Fee, _ = strconv.ParseFloat(mChannelRate, 64)
			//		payType.HandlingFee, _ = strconv.ParseFloat(mChannelHandingFee, 64)
			//		payType.DailyTxLimit, _ = strconv.ParseFloat(mOneDayTransactionLimit, 64)
			//		payType.SingleMinCharge, _ = strconv.ParseFloat(mSingleLimitMinimum, 64)
			//		payType.SingleMaxCharge, _ = strconv.ParseFloat(mSingleLimitMaximum, 64)
			//		payType.MaxInternalCharge, _ = strconv.ParseFloat(mMaximunAmountOfCharge, 64)
			//		payType.IsProxy = method.WalletShareType
			//		payType.FixedAmount = method.FixedAmount
			//		payType.Device = method.Device
			//		payType.Status = method.ChannelEnableStatus
			//		if err := tx.Table("ch_channel_pay_types").
			//			Where("id = ?", payType.ID).
			//			Updates(map[string]interface{}{"fee": payType.Fee,
			//				"handling_fee": payType.HandlingFee,
			//				"daily_tx_limit": payType.DailyTxLimit,
			//				"single_min_charge": payType.SingleMinCharge,
			//				"single_max_charge": payType.SingleMaxCharge,
			//				"max_internal_charge": payType.MaxInternalCharge,
			//				"is_proxy": payType.IsProxy,
			//				"fixed_amount": payType.FixedAmount,
			//				"device": payType.Device,
			//				"status": payType.Status,
			//			}).Error; err != nil {
			//			return errorz.New(response.DATABASE_FAILURE, err.Error())
			//		}
			//	}
			//	//else {
			//	//	channelPayTypeUpdate = append(channelPayTypeUpdate, payType)
			//	//}
			//}
		}

		if err := tx.Table("ch_channel_pay_types").CreateInBatches(channelPayTypeCreates, len(channelPayTypeCreates)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		//if err := tx.Table("ch_channel_pay_types").Updates(&channelPayTypeUpdate).Error; err != nil {
		//	return errorz.New(response.DATABASE_FAILURE, err.Error())
		//}
		return nil
	})
}

func (l *ChannelEtlLogic) payType(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var payTypes []PayType
		if err := oldDb.Table("param_data").Where("function_name = 'channel' AND param_type = 'payType'").Find(&payTypes).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var payTypesV2 []types.PayType
		if err := tx.Table("ch_pay_types").Find(&payTypesV2).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var newPayTypes []types.PayType

		var payTypeMap map[string]types.PayType
		payTypeMap = make(map[string]types.PayType)
		for _, t := range payTypesV2 {
			payTypeMap[t.Code] = t
		}
		for _, payType := range payTypes {

			if _, ok := payTypeMap[payType.ParamNumber]; !ok {
				newPayType := types.PayType{
					Code:     payType.ParamNumber,
					Name:     payType.ParamName,
					Currency: "CNY",
				}
				newPayTypes = append(newPayTypes, newPayType)
			}
		}

		if err := tx.Table("ch_pay_types").CreateInBatches(newPayTypes, len(newPayTypes)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *ChannelEtlLogic) updateChannelIsWalletShare(oldDB *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var channelDatas []ChannelData
		if err := oldDB.Table("channel_data").Find(&channelDatas).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var channels []types.ChannelData
		if err := tx.Table("ch_channels").Find(&channels).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		channelMap := make(map[string]types.ChannelData)
		for _, channel := range channels {
			channelMap[channel.Code] = channel
		}

		for _, data := range channelDatas {
			if v, ok := channelMap[data.ChannelCoding]; ok {
				v.IsProxy = data.IsWalletShare
				if err := tx.Table("ch_channels").Updates(&v).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			}
		}
		return nil
	})
}

func (l *ChannelEtlLogic) Banks(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var bankMains []BankMain
		if err := oldDb.Table("bank_main").Where("status = '1'").Find(&bankMains).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var banks []types.Bank
		if err := tx.Table("bk_banks").Find(&banks).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		bankMap := make(map[string]types.Bank)
		for _, bank := range banks {
			bankMap[bank.BankNo] = bank
		}

		var newBanks []types.Bank
		for _, main := range bankMains {
			if v, ok := bankMap[main.MainBankNo]; ok {
				v.BankNameEn = main.EngName
				v.Abbr = main.Abbreviation
				if err := tx.Table("bk_banks").Updates(v).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			} else {
				newBank := types.Bank{
					BankNo:       main.MainBankNo,
					BankName:     main.MainBankName,
					BankNameEn:   main.EngName,
					Abbr:         main.Abbreviation,
					CurrencyCode: main.CurrencyCoding,
					Status:       "1",
				}
				newBanks = append(newBanks, newBank)
			}
		}
		if err := tx.Table("bk_banks").CreateInBatches(newBanks, len(newBanks)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *ChannelEtlLogic) channelBank(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var bankChannels []BankChannel
		if err := oldDb.Table("bank_channel_map").Find(&bankChannels).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var channels []types.ChannelData
		if err := tx.Table("ch_channels").Find(&channels).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		channelMap := make(map[string]string)
		for _, channel := range channels {
			channelMap[channel.Code] = channel.Code
		}

		//channelBankMaps := make(map[string][]types.ChannelBank)
		for k := range channelMap {
			var channelBanks []types.ChannelBank
			for _, channel := range bankChannels {
				if k == channel.ChannelCoding {
					channelBank := types.ChannelBank{
						ChannelCode: k,
						BankNo:      channel.MainBankNo,
						MapCode:     channel.BankCode,
					}
					channelBanks = append(channelBanks, channelBank)
				}
			}
			if err := tx.Table("ch_channel_banks").CreateInBatches(channelBanks, len(channelBanks)).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		return nil
	})
}
