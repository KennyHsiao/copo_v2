package etl

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"database/sql"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type MerchantEtlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}
type MerchantData struct {
	MerchantCoding                         string `json:"merchant_coding"`
	MerchantKey                            string `json:"merchant_key"`
	MerchantGrade                          string `json:"merchant_grade"`
	AccountActivationStatus                string `json:"account_activation_status"`
	MerchantPhoneNumber                    string `json:"merchant_phone_number"`
	MerchantMailBox                        string `json:"merchant_mail_Box"`
	MerchantCommunicationSoftware          string `json:"merchant_communication_software"`
	MerchantCommunicationSoftwareNickname  string `json:"merchant_communication_software_nickname"`
	MerchantCommunicationSoftwareGroupName string `json:"merchant_communication_software_group_name"`
	MerchantCommunicationGroupId           string `json:"merchant_communication_group_id"`
	MerchantCompanyName                    string `json:"merchant_company_name"`
	MerchantOperatingWebsite               string `json:"merchant_operating_website"`
	MerchantTestAccount                    string `json:"merchant_test_account"`
	MerchantTestPassword                   string `json:"merchant_test_password"`
	WithdrawSetting                        string `json:"withdraw_setting"`
	WithdrawType                           string `json:"withdraw_type"`
	ChannelUseType                         string `json:"channel_use_type"`
	IsAgentRole                            string `json:"is_agent_role"`
	IsHaveAgentMerchant                    string `json:"is_have_agent_merchant"`
	ParentMerchantCoding                   string `json:"parent_merchant_coding"`
	AgentLayerNo                           string `json:"agent_layer_no"`
	AccountName                            string `json:"account_name"`
	Password                               string `json:"password"`
}

type UserMerchant struct {
	UserAccount  string `json:"user_account"`
	MerchantCode string `json:"merchant_code"`
}

type WhiteList struct {
	MerchantCoding string `json:"merchant_coding"`
	OpenIp         string `json:"open_ip"`
	Type           string `json:"type"`
}

type MerchantWhiteList struct {
	MerchantCode string `json:"merchant_code"`
	BoIp         string `json:"bo_ip"`
	ApiIp        string `json:"api_ip"`
}

type MerchantCurrency struct {
	MerchantCoding string `json:"merchant_coding"`
	CurrencyCoding string `json:"currency_coding"`
	IsUse          string `json:"is_use"`
}

type MerchantChannelRateData struct {
	MerchantCoding      string  `json:"merchant_coding"`
	PayTypeCoding       string  `json:"pay_type_coding"`
	MerchantRate        float64 `json:"merchant_rate"`
	MerchantHandlingFee string  `json:"merchant_handling_fee"`
	Designation         string  `json:"designation"`
	PayTypeSubCoding    string  `json:"pay_type_sub_coding"`
}

type MerchantWalletData struct {
	MerchantCoding    string  `json:"merchant_coding"`
	Currency          string  `json:"currency"`
	PayFrozenAmount   float64 `json:"pay_frozen_amount"`
	ExchangeAmount    float64 `json:"exchange_amount"`
	ProxyFrozenAmount float64 `json:"proxy_frozen_amount"`
	TotalAmount       float64 `json:"total_amount"`
	PaySum            float64 `json:"pay_sum"`
	DelaySum          float64 `json:"delay_sum"`
	WalletShareAmount float64 `json:"wallet_share_amount"`
}

type MerchantCommissionWallet struct {
	MerchantCoding string  `json:"merchant_coding"`
	CurrencyCoding string  `json:"currency_coding"`
	Money          float64 `json:"money"`
}

type MerchantBindingBankData struct {
	MerchantCoding string `json:"merchant_coding"`
	CurrencyCoding string `json:"currency_coding"`
	Province       string `json:"province"`
	City           string `json:"city"`
	BankName       string `json:"bank_name"`
	AccountNo      string `json:"account_no"`
	PayeeName      string `json:"payee_name"`
}

type AgentRecord struct {
	MerchantCoding       string `json:"merchant_coding"`
	ParentMerchantCoding string `json:"parent_merchant_coding"`
	AgentLayerNo         string `json:"agent_layer_no"`
}

type McAgentRecord struct {
	MerchantCode         string    `json:"merchant_code"`
	AgentParentCode      string    `json:"agent_parent_code"`
	AgentLayerCode       string    `json:"agent_layer_code"`
	ParentAgentLayerCode string    `json:"parent_agent_layer_code"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func NewMerchantEtlLogic(ctx context.Context, svcCtx *svc.ServiceContext) MerchantEtlLogic {
	return MerchantEtlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MerchantEtlLogic) MerchantEtl(req *types.MerchantRequest) error {
	//oldDb, err := mysqlz.New("8.129.209.41", "3306", "dior", "P#tjnnPEZ@JwQjkFrcdG", "dior22").
	//	SetCharset("utf8mb4").
	//	SetLoc("UTC").
	//	SetLogger(logrusz.New().SetLevel("debug").Writer()).
	//	Connect(mysqlz.Pool(1, 1, 1))
	//if err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, err.Error())
	//}

	//// 主號戶導入
	//if err := l.merchantNotSubAccountAndUserMap(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "主帳戶導入錯誤："+err.Error())
	//}
	//// 子帳戶導入
	//if err := l.merchantSubAccountAndUserMap(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "子帳戶導入錯誤："+err.Error())
	//}
	// 商戶管理員權限導入
	//if err := l.MerchantAccountROle(); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商戶管理員權限導入錯誤："+err.Error())
	//}
	// 代理商户管理员权限导入
	if err := l.MerchantAgentRole(); err != nil {
		return errorz.New(response.DATABASE_FAILURE, "代理商户管理员权限导入错误："+err.Error())
	}
	// 商戶白名單導入
	//if err := l.merchantWhiteList(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商戶白名單導入錯誤："+err.Error())
	//}
	// 商戶幣別與錢包
	//if err := l.merchantCurrencyAndBalances(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商戶幣別與錢包錯誤："+err.Error())
	//}
	// 商戶費率與支付類型
	//if err := l.merchantRateAndPayType(oldDb); err != nil{
	//	return errorz.New(response.DATABASE_FAILURE, "商戶費率與支付類型錯誤："+err.Error())
	//}
	// 商戶常用帳戶
	//if err := l.merchantOftenAccount(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商戶常用帳戶導入錯誤 ："+ err.Error())
	//}
	// 商户代理层级编码记录
	//if err := l.merchantAgentRecord(oldDb); err != nil{
	//	return errorz.New(response.DATABASE_FAILURE, "商户代理层级编码记录 :"+ err.Error())
	//}
	//商户钱包导入(下发钱包 带付钱包 )
	//if err := l.merchantBalances(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商户钱包导入(下发钱包 带付钱包 )錯誤："+err.Error())
	//}
	//商户钱包导入(佣金钱包) (not use)
	//if err := l.merchantCommissionBalances(oldDb); err != nil {
	//	return errorz.New(response.DATABASE_FAILURE, "商户钱包导入(佣金钱包)錯誤："+err.Error())
	//}
	return nil
}

func (l *MerchantEtlLogic) merchantNotSubAccountAndUserMap(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var merchants []MerchantData

		selectX := "a.merchant_coding as MerchantCoding, a.merchant_key as MerchantKey, a.merchant_grade as MerchantGrade, a.account_activation_status as AccountActivationStatus," +
			"a.merchant_phone_number as MerchantPhoneNumber, a.merchant_mailbox MerchantMailBox, a.merchant_communication_software as MerchantCommunicationSoftware, a.merchant_communication_software_nickname as MerchantCommunicationNickname," +
			"a.merchant_communication_software_group_name as MerchantCommunicationSoftwareGroupName, a.merchant_communication_group_id as MerchantCommunicationGroupId, a.merchant_operating_website as MerchantOperatingWebsite, a.merchant_test_account as MerchantTestAccount," +
			"a.merchant_test_password as MerchantTestPassword, a.withdraw_setting as WithdrawSetting, a.withdraw_type as WithdrawType, a.channel_use_type as ChannelUseType," +
			"a.is_agent_role as IsAgentRole, a.is_have_agent_merchant as IsHaveAgentMerchant, a.parent_merchant_coding as ParentMerchantCoding, a.agent_layer_no as AgentLayerNo," +
			"a.merchant_company_name as MerchantCompanyName, a.password as Password, b.account_name as AccountName"
		if err := oldDb.Table("merchant_data a").
			Select(selectX).
			Joins("join account b on a.merchant_coding = b.merchant_coding").
			Where("b.is_subject_account != '1'").Find(&merchants).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var merchantXs []types.MerchantX
		var userMerchants []UserMerchant
		//var accounts []types.User
		for _, record := range merchants {
			merchantContact := types.MerchantContact{
				Phone:                 record.MerchantPhoneNumber,
				Email:                 record.MerchantMailBox,
				CommunicationSoftware: record.MerchantCommunicationSoftware,
				CommunicationNickname: record.MerchantCommunicationSoftwareNickname,
				GroupName:             record.MerchantCommunicationSoftwareGroupName,
				GroupID:               record.MerchantCommunicationGroupId,
			}
			merchantBizInfo := types.MerchantBizInfo{
				CompanyName:      record.MerchantCompanyName,
				OperatingWebsite: record.MerchantOperatingWebsite,
				TestAccount:      record.MerchantTestAccount,
				TestPassword:     record.MerchantTestPassword,
			}
			billLadingType := "0"
			if record.ChannelUseType == "2" {
				billLadingType = "1"
			}
			agentParentCode := ""
			if record.IsHaveAgentMerchant == "1" {
				agentParentCode = record.ParentMerchantCoding
			}
			payingValidatedType := ""
			if len(record.WithdrawType) > 0 {
				payingValidatedType = record.WithdrawType
			}

			regTime := time.Now().UTC().Unix()
			merchant := types.Merchant{
				Code:                record.MerchantCoding,
				ScrectKey:           record.MerchantKey,
				BillLadingType:      billLadingType,
				Status:              record.AccountActivationStatus,
				AgentStatus:         record.IsAgentRole,
				AgentLayerCode:      record.AgentLayerNo,
				AgentParentCode:     agentParentCode,
				BizInfo:             merchantBizInfo,
				Contact:             merchantContact,
				LoginValidatedType:  "1",
				PayingValidatedType: payingValidatedType,
				WithdrawPassword:    record.Password,
				IsWithdraw:          record.WithdrawSetting,
				RegisteredAt:        regTime,
			}
			merchantX := types.MerchantX{
				Merchant: merchant,
			}
			merchantXs = append(merchantXs, merchantX)

			userMerchant := UserMerchant{
				UserAccount:  record.AccountName,
				MerchantCode: record.MerchantCoding,
			}

			//user := types.User{
			//	Account: record.AccountName,
			//	Email: record.MerchantMailBox,
			//	Phone: record.MerchantPhoneNumber,
			//	DisableDelete: "1",
			//}
			userMerchants = append(userMerchants, userMerchant)
			////accounts = append(accounts, user)
			if err := tx.Table("au_users").Where("account = ?", record.AccountName).
				Updates(map[string]interface{}{"disable_delete": "1", "email": record.MerchantMailBox, "phone": record.MerchantPhoneNumber}).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}
		if err := tx.Table("mc_merchants").CreateInBatches(merchantXs, len(merchantXs)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		if err := tx.Table("au_user_merchants").CreateInBatches(userMerchants, len(userMerchants)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *MerchantEtlLogic) merchantSubAccountAndUserMap(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var merchants []MerchantData

		selectX := "a.merchant_coding as MerchantCoding, a.merchant_key as MerchantKey, a.merchant_grade as MerchantGrade, a.account_activation_status as AccountActivationStatus," +
			"a.merchant_phone_number as MerchantPhoneNumber, a.merchant_mailbox MerchantMailBox, a.merchant_communication_software as MerchantCommunicationSoftware, a.merchant_communication_software_nickname as MerchantCommunicationNickname," +
			"a.merchant_communication_software_group_name as MerchantCommunicationSoftwareGroupName, a.merchant_communication_group_id as MerchantCommunicationGroupId, a.merchant_operating_website as MerchantOperatingWebsite, a.merchant_test_account as MerchantTestAccount," +
			"a.merchant_test_password as MerchantTestPassword, a.withdraw_setting as WithdrawSetting, a.withdraw_type as WithdrawType, a.channel_use_type as ChannelUseType," +
			"a.is_agent_role as IsAgentRole, a.is_have_agent_merchant as IsHaveAgentMerchant, a.parent_merchant_coding as ParentMerchantCoding, a.agent_layer_no as AgentLayerNo," +
			"a.merchant_company_name as MerchantCompanyName, a.password as Password, b.account_name as AccountName"
		if err := oldDb.Table("merchant_data a").
			Select(selectX).
			Joins("join account b on a.merchant_coding = b.merchant_coding").
			Where("b.is_subject_account = '1'").Find(&merchants).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var userMerchants []UserMerchant
		//var accounts []types.User
		for _, record := range merchants {
			userMerchant := UserMerchant{
				UserAccount:  record.AccountName,
				MerchantCode: record.MerchantCoding,
			}
			//user := types.User{
			//	Account: record.AccountName,
			//	Email: record.MerchantMailBox,
			//	Phone: record.MerchantPhoneNumber,
			//	DisableDelete: "1",
			//}
			userMerchants = append(userMerchants, userMerchant)
			//accounts = append(accounts, user)
			if err := tx.Table("au_users").Where("account = ?", record.AccountName).
				Updates(map[string]interface{}{"email": record.MerchantMailBox, "phone": record.MerchantPhoneNumber}).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		//if err := tx.Table("au_users_copy1").Updates(accounts).Error; err != nil{
		//	return errorz.New(response.DATABASE_FAILURE, err.Error())
		//}
		if err := tx.Table("au_user_merchants").CreateInBatches(userMerchants, len(userMerchants)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *MerchantEtlLogic) merchantWhiteList(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var whiteLists []WhiteList
		if err := oldDb.Table("merchant_white_list").Find(&whiteLists).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var apiWhiteLists []WhiteList
		var boWhiteLists []WhiteList
		merchantMap := make(map[string]string)
		for _, list := range whiteLists {
			if list.Type == "0" {
				apiWhiteLists = append(apiWhiteLists, list)
			} else {
				boWhiteLists = append(boWhiteLists, list)
			}
			merchantMap[list.MerchantCoding] = list.MerchantCoding
		}
		//var merchantWhiteLists []MerchantWhiteList
		for k, _ := range merchantMap {
			var apiIps string
			var boIps string
			for _, list := range apiWhiteLists {
				if list.MerchantCoding == k {
					//if (len(apiWhiteLists) -1) == i {
					//	apiIps = apiIps + list.OpenIp
					//}else {
					//	apiIps = apiIps + list.OpenIp + ","
					//}
					apiIps = apiIps + list.OpenIp + ","
				}
			}
			for _, list := range boWhiteLists {
				if list.MerchantCoding == k {
					//if (len(boWhiteLists) -1) == i {
					//	boIps = boIps + list.OpenIp
					//}else {
					//	boIps = boIps + list.OpenIp + ","
					//}
					boIps = boIps + list.OpenIp + ","
				}
			}
			//merchantWhiteList := MerchantWhiteList{
			//	MerchantCode: k,
			//	ApiIp: apiIps,
			//	BoIp: boIps,
			//}
			//merchantWhiteLists = append(merchantWhiteLists, merchantWhiteList)
			apiIps = strings.TrimRight(apiIps, ",")
			boIps = strings.TrimRight(boIps, ",")
			if err := tx.Table("mc_merchants").Where("code = ?", k).
				Updates(map[string]interface{}{"bo_ip": boIps, "api_ip": apiIps}).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
		}

		return nil
	})
}

func (l *MerchantEtlLogic) merchantCurrencyAndBalances(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var merchantCurrencies []MerchantCurrency
		if err := oldDb.Table("merchant_currency_data").Find(&merchantCurrencies).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var merchantCurrencies2 []types.MerchantCurrency
		for _, currency := range merchantCurrencies {
			var merchantCurrency types.MerchantCurrency
			if (currency.IsUse != "0" && currency.CurrencyCoding == "CNY") || (currency.IsUse != "0" && currency.CurrencyCoding == "THB") ||
				(currency.IsUse != "0" && currency.CurrencyCoding == "USD") || (currency.IsUse != "0" && currency.CurrencyCoding == "USDT") {
				merchantCurrency.MerchantCode = currency.MerchantCoding
				merchantCurrency.CurrencyCode = currency.CurrencyCoding
				if currency.IsUse == "2" {
					merchantCurrency.Status = "0"
				} else if currency.IsUse == "1" {
					merchantCurrency.Status = "1"
				}
				merchantCurrencies2 = append(merchantCurrencies2, merchantCurrency)
			}
		}
		if err := tx.Table("mc_merchant_currencies").CreateInBatches(merchantCurrencies2, len(merchantCurrencies2)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var merchantBalances []types.MerchantBalanceX
		for _, currency := range merchantCurrencies2 {
			var merchantBalances1 types.MerchantBalanceX
			var merchantBalances2 types.MerchantBalanceX
			var merchantBalances3 types.MerchantBalanceX
			merchantBalances1.MerchantCode = currency.MerchantCode
			merchantBalances1.CurrencyCode = currency.CurrencyCode
			merchantBalances1.BalanceType = constants.DF_BALANCE
			merchantBalances2.MerchantCode = currency.MerchantCode
			merchantBalances2.CurrencyCode = currency.CurrencyCode
			merchantBalances2.BalanceType = constants.XF_BALANCE
			merchantBalances3.MerchantCode = currency.MerchantCode
			merchantBalances3.CurrencyCode = currency.CurrencyCode
			merchantBalances3.BalanceType = constants.YJ_BALANCE
			merchantBalances = append(merchantBalances, merchantBalances1)
			merchantBalances = append(merchantBalances, merchantBalances2)
			merchantBalances = append(merchantBalances, merchantBalances3)
		}

		if err := tx.Table("mc_merchant_balances").CreateInBatches(merchantBalances, len(merchantBalances)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantEtlLogic) merchantRateAndPayType(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var channelPayTypes []types.ChannelPayType
		if err := tx.Table("ch_channel_pay_types").Find(&channelPayTypes).Where("status IN ('1','2')").Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var merchantChannelRateDatas []MerchantChannelRateData
		var merchantChannelRates []types.MerchantChannelRate
		for _, payType := range channelPayTypes {

			if payType.Status != "0" {
				merchantChannelRateData := []MerchantChannelRateData{}
				if err := oldDb.Table("merchant_channel_rate_data").Where("pay_type_coding = ?", payType.Code).Find(&merchantChannelRateData).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
				merchantChannelRateDatas = append(merchantChannelRateDatas, merchantChannelRateData...)
			}
		}
		var merchants []types.Merchant
		if err := tx.Table("mc_merchants").Find(&merchants).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		merchantMap := make(map[string]string)
		for _, s := range merchants {
			merchantMap[s.Code] = s.Code
		}

		for _, rate := range merchantChannelRateDatas {
			if _, ok := merchantMap[rate.MerchantCoding]; ok {
				oldMerchantChannelRate := []types.MerchantChannelRate{}
				if err := tx.Table("mc_merchant_channel_rate").
					Where("merchant_code = ?", rate.MerchantCoding).
					Find(&oldMerchantChannelRate).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
				oldMerchantChannelRateMap := make(map[string]string)
				for _, s := range oldMerchantChannelRate {
					oldMerchantChannelRateMap[s.ChannelPayTypesCode] = s.ChannelPayTypesCode
				}

				if _, ok2 := oldMerchantChannelRateMap[rate.PayTypeCoding]; !ok2 {
					payTypeCoding := rate.PayTypeCoding
					channelCode := payTypeCoding[:9]
					payTypeCode := payTypeCoding[9:]
					handlingFee, _ := strconv.ParseFloat(rate.MerchantHandlingFee, 64)
					merchantChannelRate := types.MerchantChannelRate{
						MerchantCode:        rate.MerchantCoding,
						ChannelCode:         channelCode,
						PayTypeCode:         payTypeCode,
						ChannelPayTypesCode: rate.PayTypeCoding,
						Fee:                 rate.MerchantRate,
						HandlingFee:         handlingFee,
						Status:              "1",
						Designation:         rate.Designation,
						DesignationNo:       rate.PayTypeSubCoding,
					}
					merchantChannelRates = append(merchantChannelRates, merchantChannelRate)
				}
			}
		}
		if err := tx.Table("mc_merchant_channel_rate").CreateInBatches(merchantChannelRates, len(merchantChannelRates)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantEtlLogic) merchantBalances(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		selectX := "with dsm as (" +
			" select merchant_coding, sum(ifnull(p.pay_amount, 0))-sum(ifnull(p.charge_fee, 0)) pay_amount,c.currency_coding currency" +
			" from merchant_pay_order p" +
			" left join channel_data c on c.channel_coding = SUBSTR(p.pay_type_coding,1,9)" +
			" where order_status in (1, 3)" +
			" and (confirm_time is not null or call_back_time is not null)" +
			" and DATE_ADD(STR_TO_DATE(IFNULL(confirm_time, call_back_time), '%Y%m%d'),INTERVAL delay_pay_days DAY ) > CURDATE()" +
			" and merchant_coding= @merchantCode" +
			" group by currency )," +

			" exo as ( select merchant_coding,c.currency_coding currency," +
			" sum(case when exchange_amount is not null and exchange_amount != '' then exchange_amount else 0 end) exchange_amount" +
			" from merchant_exchange_order meo" +
			" left join channel_data c on c.channel_coding = meo.channel_coding" +
			" where order_status='0'" +
			" and merchant_coding= @merchantCode" +
			" group by c.currency_coding)," +

			" sum as ( select w.merchant_coding,c.currency_coding currency," +
			" sum(money) total_amount," +
			" sum(case when SUBSTR(w.pay_type_coding,10,2) !='DF' then money else 0 end) pay_sum," +
			" sum(case when p.wallet_share_type ='1' then money else 0 end) wallet_share_amount," +
			" md.agent_layer_no" +
			" from merchant_wallet_data w" +
			" left join channel_pay_method p on p.pay_type_coding=w.pay_type_coding" +
			" left join channel_data c on c.channel_coding = SUBSTR(w.pay_type_coding,1,9)" +
			" left join merchant_data   md on w.merchant_coding = md.merchant_coding" +
			" where w.merchant_coding= @merchantCode" +
			" group by c.currency_coding)" +

			" select sum.merchant_coding, sum.currency, sum(case when pfro.type='1' then pfro.amount else 0 end) as pay_frozen_amount," +
			" sum(case when exo.exchange_amount is not null and exo.exchange_amount != '' then exo.exchange_amount else 0 end) exchange_amount," +
			" sum(case when pfro.type='2' then pfro.amount else 0 end) as proxy_frozen_amount," +
			" total_amount, pay_sum, IFNULL(dsm.pay_amount,0) delay_sum, IFNULL(wallet_share_amount,0) wallet_share_amount, agent_layer_no" +
			" from sum" +
			" left join exo on exo.currency = sum.currency and exo.merchant_coding = sum.merchant_coding" +
			" left outer join dsm on dsm.currency = sum.currency and dsm.merchant_coding = sum.merchant_coding" +
			" left outer join merchant_frozen_data pfro on sum.merchant_coding = pfro.merchant_coding and pfro.currency = sum.currency" +
			" where sum.currency= @currencyCode" +
			" group by sum.currency"

		var merchantBalances []types.MerchantBalance
		if err := tx.Table("mc_merchant_balances").Find(&merchantBalances).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		for _, balance := range merchantBalances {
			var merchantWalletData MerchantWalletData
			if err := oldDb.Raw(selectX, sql.Named("merchantCode", balance.MerchantCode), sql.Named("currencyCode", balance.CurrencyCode)).
				Find(&merchantWalletData).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			proxyPayAmount := utils.FloatSub(merchantWalletData.TotalAmount, merchantWalletData.PaySum)
			realProxyAmount := utils.FloatSub(proxyPayAmount, merchantWalletData.ProxyFrozenAmount)
			merchantDFB := types.MerchantBalance{
				Balance: realProxyAmount,
			}
			if err := tx.Table("mc_merchant_balances").
				Where("merchant_code = ?", balance.MerchantCode).
				Where("currency_code = ?", balance.CurrencyCode).
				Where("balance_type = ?", constants.DF_BALANCE).
				Updates(&merchantDFB).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}

			payAmount := utils.FloatSub(merchantWalletData.PaySum, merchantWalletData.PayFrozenAmount)
			merchantXFB := types.MerchantBalance{
				Balance: payAmount,
			}

			if err := tx.Table("mc_merchant_balances").
				Where("merchant_code = ?", balance.MerchantCode).
				Where("currency_code = ?", balance.CurrencyCode).
				Where("balance_type = ?", constants.XF_BALANCE).
				Updates(merchantXFB).Error; err != nil {
				return errorz.New(response.DATABASE_FAILURE, err.Error())
			}
			//var merchantBalances2 [] types.MerchantBalance
			//merchantBalances2 = append(merchantBalances2, merchantDFB)
			//merchantBalances2 = append(merchantBalances2, merchantXFB)
			//if err := tx.Table("mc_merchant_balances_copy1").
			//	Where("merchant_code = ?", balance.MerchantCode).
			//	Where("currency_code = ?", balance.CurrencyCode).Updates(&merchantBalances2).Error; err != nil {
			//	return errorz.New(response.DATABASE_FAILURE, err.Error())
			//}
		}

		return nil
	})
}

//func (l *MerchantEtlLogic) merchantCommissionBalances(oldDb *gorm.DB) error {
//	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
//		var merchantBalances []types.MerchantBalance
//		if err := tx.Table("mc_merchant_balances").Find(&merchantBalances).Error; err != nil {
//			return errorz.New(response.DATABASE_FAILURE, err.Error())
//		}
//		for _, balance := range merchantBalances {
//			var merchantCommissionWallet MerchantCommissionWallet
//			if err := oldDb.Table("merchant_commission_wallet").
//				Where("merchant_coding = ?", balance.MerchantCode).
//				Where("currency_coding = ?", balance.CurrencyCode).
//				Find(&merchantCommissionWallet).Error; err != nil {
//				return errorz.New(response.DATABASE_FAILURE, err.Error())
//			}
//			merchantYJB := types.MerchantBalance{
//				Balance: merchantCommissionWallet.Money,
//			}
//			if err := tx.Table("mc_merchant_balances").
//				Where("merchant_code = ?", balance.MerchantCode).
//				Where("currency_code = ?", balance.CurrencyCode).
//				Where("balance_type = ?", constants.YJ_BALANCE).
//				Updates(&merchantYJB).Error; err != nil {
//				return errorz.New(response.DATABASE_FAILURE, err.Error())
//			}
//		}
//		return nil
//	})
//}

func (l *MerchantEtlLogic) MerchantAccountROle() error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var users []types.User
		if err := tx.Table("au_users").Where("is_admin != '1' and disable_delete = 1").Find(&users).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var userRoles []UserRole
		for _, user := range users {
			userRole := UserRole{
				UserId: user.ID,
				RoleId: 2,
			}
			userRoles = append(userRoles, userRole)
		}
		if err := l.svcCtx.MyDB.Table("au_user_roles").CreateInBatches(userRoles, len(userRoles)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantEtlLogic) MerchantSubAccountRole() error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var users []types.User
		if err := tx.Table("au_users").Where("is_admin != '1' and disable_delete = 0").Find(&users).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		var userRoles []UserRole
		for _, user := range users {
			userRole := UserRole{
				UserId: user.ID,
				RoleId: 131,
			}
			userRoles = append(userRoles, userRole)
		}
		if err := l.svcCtx.MyDB.Table("au_user_roles").CreateInBatches(userRoles, len(userRoles)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantEtlLogic) merchantOftenAccount(oldDB *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var merchantBindingBankData []MerchantBindingBankData
		if err := oldDB.Table("merchant_binding_bank_data").Where("status = '1'").Find(&merchantBindingBankData).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var merchantBindBanks []types.MerchantBindBank
		for _, datum := range merchantBindingBankData {
			merchantBindBank := types.MerchantBindBank{
				MerchantCode: datum.MerchantCoding,
				CurrencyCode: datum.CurrencyCoding,
				Province:     datum.Province,
				City:         datum.City,
				BankAccount:  datum.AccountNo,
				BankName:     datum.BankName,
				Name:         datum.PayeeName,
			}
			merchantBindBanks = append(merchantBindBanks, merchantBindBank)
		}
		if err := tx.Table("mc_merchant_bind_bank").CreateInBatches(merchantBindBanks, len(merchantBindBanks)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}
		return nil
	})
}

func (l *MerchantEtlLogic) merchantAgentRecord(oldDb *gorm.DB) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var agentRecordDatas []AgentRecord
		if err := oldDb.Distinct("merchant_coding", "parent_merchant_coding", "agent_layer_no").
			Table("agent_record").
			Where("parent_merchant_coding != '00000000'").Find(&agentRecordDatas).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		var agentRecords []McAgentRecord
		for _, datum := range agentRecordDatas {
			parentLayerCode := datum.ParentMerchantCoding[len(datum.ParentMerchantCoding)-4:]
			agentRecord := McAgentRecord{
				MerchantCode:         datum.MerchantCoding,
				AgentParentCode:      datum.ParentMerchantCoding,
				ParentAgentLayerCode: parentLayerCode,
				AgentLayerCode:       datum.AgentLayerNo,
			}
			agentRecords = append(agentRecords, agentRecord)
		}
		if err := tx.Table("mc_agent_record").CreateInBatches(agentRecords, len(agentRecords)).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *MerchantEtlLogic) MerchantAgentRole() error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		var users []types.User
		if err := tx.Table("au_users").
			Preload("Merchants").
			Where("is_admin != '1' and disable_delete = 1").Find(&users).Error; err != nil {
			return errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		for _, user := range users {
			merchant := user.Merchants[0]
			if merchant.AgentStatus == "1" {
				var userRole UserRole
				if err := tx.Table("au_user_roles").Where("user_id = ?", user.ID).Find(&userRole).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
				userRole.RoleId = 3
				if err := tx.Table("au_user_roles").Where("user_id = ?", user.ID).Updates(userRole).Error; err != nil {
					return errorz.New(response.DATABASE_FAILURE, err.Error())
				}
			}
		}

		return nil
	})
}
