package channeldata

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
)

type ChannelDataCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelDataCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelDataCreateLogic {
	return ChannelDataCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelDataCreateLogic) ChannelDataCreate(req types.ChannelDataCreateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) error {
		channelData := &types.ChannelDataCreate{
			ChannelDataCreateRequest: req,
		}

		//检查请求参数
		paramErr := model.CheckRequestCreateValue(req)
		if paramErr != nil {
			return paramErr
		}
		//将最新渠道的channleCode加 1
		channelData.Code = getNewChannelCode(db)
		//设定预设值
		setDefaultValue(req, channelData)

		//新增channel
		channelInsertErr := db.Table("ch_channels").Omit("ChannelPayTypeList").Create(channelData).Error
		if channelInsertErr != nil {
			return channelInsertErr
		}

		//新增渠道支付方式
		if len(req.ChannelPayTypeList) > 0 {
			var channelPayTypeList []types.ChannelPayTypeCreateRequest
			//datas := make(map[string]string)
			for _, m := range req.ChannelPayTypeList {
				//datas[Map] = Map.MapCode
				//同步新增到channel_pay_type
				createChannelPayType := types.ChannelPayTypeCreateRequest{
					Code:              channelData.Code + m.PayTypeCode, //渠道號+支付代碼
					ChannelCode:       channelData.Code,
					PayTypeCode:       m.PayTypeCode,
					Status:            constants.OFF,
					MapCode:           m.MapCode,
					Fee:               m.Fee,
					HandlingFee:       m.HandlingFee,
					MaxInternalCharge: m.MaxInternalCharge,
					DailyTxLimit:      m.DailyTxLimit,
					SingleMaxCharge:   m.SingleMaxCharge,
					SingleMinCharge:   m.SingleMinCharge,
					FixedAmount:       m.FixedAmount,
					BillDate:          m.BillDate,
					IsProxy:           m.IsProxy,
					IsRate:            "0", //是否費率計算 default: 0: No
					Device:            "All",
				}
				//insertErr := model.NewChannelPayType(db).SingleInsertChannelPayType(createChannelPayType)
				channelPayTypeList = append(channelPayTypeList, createChannelPayType)
			}
			if len(channelPayTypeList) > 0 {
				insertErr := model.NewChannelPayType(db).InsertChannelPayType(channelPayTypeList)
				if insertErr != nil {
					return insertErr
				}
			}

			//p, _ := json.Marshal(datas)
			//channelData.PayTypeMap = string(p)
			logx.Info("channelData.PayTypeMap: ", channelData.PayTypeMap)
		}

		//新增渠道銀行
		if len(req.BankCodeMapList) > 0 {
			var channelBankList []types.ChannelBankCreateRequest
			for _, m := range req.BankCodeMapList {
				channelBank := types.ChannelBankCreateRequest{
					ChannelCode: channelData.Code,
					BankNo:      m.BankNo,
					MapCode:     m.MapCode,
				}
				channelBankList = append(channelBankList, channelBank)
			}
			if err := model.NewChannelBank(db).InsertChannelBank(channelBankList); err != nil {
				return err
			}
		}
		logx.Infof("新增渠道資料: %#v", channelData)
		return nil
	})
}

func setDefaultValue(req types.ChannelDataCreateRequest, channelData *types.ChannelDataCreate) {

	if len(req.Status) < 1 {
		channelData.Status = constants.ON
	}
	//支援代付
	if len(req.IsProxy) < 1 {
		channelData.IsProxy = constants.ISPROXY
	}
	//內充預設先付
	if len(req.IsNZPre) < 1 {
		channelData.IsNZPre = constants.ISNZPRE
	}

	//設定專案名稱(沒給塞ChannelCode)
	if len(req.ProjectName) < 1 {
		channelData.ProjectName = channelData.Code
	}
}

func getNewChannelCode(db *gorm.DB) string {
	var code string
	row := db.Table("ch_channels").Select("max(code)").Row()
	row.Scan(&code)
	codeNum, _ := strconv.Atoi(code[3:len(code)])
	return "CHN" + fmt.Sprintf("%06d", codeNum+1)
}
