package order

import (
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
	"sync"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubWithdrawToChannelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubWithdrawToChannelLogic(ctx context.Context, svcCtx *svc.ServiceContext) SubWithdrawToChannelLogic {
	return SubWithdrawToChannelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubWithdrawToChannelLogic) SubWithdrawToChannel(req *types.SubWithdrawToChannelRequest) (resp string, err error) {
	order := &types.OrderX{}
	var orderChannels []types.OrderChannelsX
	var orderChannelsQuery []types.OrderChannelsX
	var totalWithdrawAmount float64
	//1. 抓取訂單2
	if err = l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", req.OrderNo).Take(&order).Error; err != nil {
		return "", errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
	}

	//2. 更新訂單change_type = 1 (下發轉代付)
	if errUpd := l.svcCtx.MyDB.Table("tx_orders").
		Where("order_no = ?", req.OrderNo).
		Updates(map[string]interface{}{"change_type": "1"}).Error; errUpd != nil {
		return "", errorz.New(response.UPDATE_FAIL, err.Error())
	}

	//2. 產生下發轉代付的單號( XFxxxxx_1、XFxxxxxx_2) => tx_order_channels
	if len(req.List) > 0 {

		if errQuery := l.svcCtx.MyDB.Table("tx_order_channels").
			Where("order_no like ?", "%"+req.OrderNo+"%").Order("order_no DESC").Limit(1).Find(&orderChannelsQuery).Error; errQuery != nil {
			return "", errorz.New(response.DATABASE_FAILURE, err.Error())
		}

		indexInt := 0
		if len(orderChannelsQuery) > 0 {
			if len(orderChannelsQuery[0].OrderNo) > 0 {
				index := orderChannelsQuery[0].OrderNo[len(orderChannelsQuery[0].OrderNo)-1:] // 2 OR 3...
				indexInt, err = strconv.Atoi(index)
			}
		}

		for i, withdralOrder := range req.List {

			if withdralOrder.WithdrawAmount > 0 { // 下发金额不得为0
				// 取得渠道下發手續費
				var channelWithdrawHandlingFee float64
				var channelPayType types.ChannelPayType

				if err1 := l.svcCtx.MyDB.Table("ch_channels").Select("channel_withdraw_charge").Where("code = ?", withdralOrder.ChannelCode).
					Take(&channelWithdrawHandlingFee).Error; err1 != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return "", errorz.New(response.DATA_NOT_FOUND)
					}
					return "", errorz.New(response.DATABASE_FAILURE, err1.Error())
				}

				if err1 := l.svcCtx.MyDB.Table("ch_channel_pay_types").Where("code = ?", withdralOrder.ChannelCode+"DF").
					Take(&channelPayType).Error; err1 != nil {
					return "", errorz.New(response.DATABASE_FAILURE, err1.Error())
				}

				// 记录下发记录
				orderChannel := types.OrderChannelsX{
					OrderChannels: types.OrderChannels{
						OrderNo:     order.OrderNo,
						OrderSubNo:  order.OrderNo + fmt.Sprintf("_%d", indexInt+i+1),
						ChannelCode: withdralOrder.ChannelCode,
						HandlingFee: channelPayType.HandlingFee, //抓channel_pay_type DF handlingFee
						Fee:         channelPayType.Fee,         //抓channel_pay_type DF fee
						OrderAmount: withdralOrder.WithdrawAmount,
						Status:      constants.PROCESSING,
						//TransferHandlingFee: 0, 交易手续费等渠道回调成功后，在平均分摊交际手续费到每张单 交易手續費(跟商户收)
					},
				}
				orderChannels = append(orderChannels, orderChannel)
				totalWithdrawAmount = utils.FloatAdd(totalWithdrawAmount, withdralOrder.WithdrawAmount)
			}
		}

		var queryOrderChannels []types.OrderChannelsX
		if err = l.svcCtx.MyDB.Table("tx_order_channels").
			Where("order_no = ?", req.OrderNo).Find(&queryOrderChannels).Error; err != nil {
			return "", errorz.New(response.ORDER_NUMBER_NOT_EXIST, err.Error())
		} else if len(queryOrderChannels) > 0 {

			for _, orderChannel := range queryOrderChannels {
				// 只有成功 跟 交易中的金额要列入金额总额
				if orderChannel.Status == "20" || orderChannel.Status == "2" {
					totalWithdrawAmount = utils.FloatAdd(totalWithdrawAmount, orderChannel.OrderAmount)
				}
			}

			// 判断渠道下发金额家总须等于订单的下发金额
			if totalWithdrawAmount != order.OrderAmount {
				return "", errorz.New(response.MERCHANT_WITHDRAW_AUDIT_ERROR)
			}

		}

		//TODO Transaction_Service create Withdraw Order
		if err1 := l.svcCtx.MyDB.Table("tx_order_channels").CreateInBatches(orderChannels, len(orderChannels)).Error; err1 != nil {
			return "", errorz.New(response.DATABASE_FAILURE, err1.Error())
		}

		//发送"下发子单"到渠道
		var errFlag = false
		var errorMsg string
		wg := &sync.WaitGroup{}
		wg.Add(len(orderChannels))
		for i, _ := range orderChannels {
			//go func() {
			if errMsg := l.doSubWithdrawChannel(order, &orderChannels[i], wg); errMsg != nil {
				logx.Errorf("渠道返回错误: %s", errMsg.Error())
				errFlag = true
				errorMsg = errMsg.Error()
			}
			//}()
		}
		//wg.Wait()
		if errFlag {
			return "", errorz.New(response.CHANNEL_REPLY_ERROR, errorMsg)
		}

	} else {
		return "", errorz.New(response.FAIL)
	}
	return "success", nil

}

func (l *SubWithdrawToChannelLogic) doSubWithdrawChannel(order *types.OrderX, o *types.OrderChannelsX, wg *sync.WaitGroup) (err error) {
	//defer wg.Done()
	logx.Infof("channelCode: %s", o.ChannelCode)

	channel := &types.ChannelData{} //多個渠道
	//抓渠道
	if err := l.svcCtx.MyDB.Table("ch_channels").Where("code = ?", o.ChannelCode).Take(&channel).Error; err != nil {
		logx.WithContext(l.ctx).Errorf("渠道不存在: %s,ChannelCode:%s", err.Error(), o.ChannelCode)
		return err
	}
	//打渠道
	proxyPayRespVO := &vo.ProxyPayRespVO{}
	var errCHN error
	proxyPayRespVO, errCHN = ordersService.CallChannel_WithdrawOrder(&l.ctx, &l.svcCtx.Config, order, o, channel)

	//更新单状态
	if errCHN != nil {
		errorMsg := fmt.Sprintf("Err:%s", errCHN.Error())
		if err := l.svcCtx.MyDB.Table("tx_order_channels").Where("order_no = ?", o.OrderNo).Updates(map[string]interface{}{"status": constants.FAIL, "error_msg": errorMsg}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("更新下发子单状态之败:%s", err.Error())
		}
		return errorz.New(errorMsg)
	} else if proxyPayRespVO.Code != "0" {
		errorMsg := fmt.Sprintf("CHN_Resp_Code:%s,CHN_Resp_Msg:%s", proxyPayRespVO.Code, proxyPayRespVO.Message)
		if err := l.svcCtx.MyDB.Table("tx_order_channels").Where("order_no = ?", o.OrderNo).Updates(map[string]interface{}{"status": constants.FAIL, "error_msg": errorMsg}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("更新下发子单状态之败:%s", err.Error())
		}
		return errorz.New(errorMsg)
	} else if proxyPayRespVO.Code == "0" {
		if err := l.svcCtx.MyDB.Table("tx_order_channels").Where("order_no = ?", o.OrderNo).Updates(map[string]interface{}{"status": constants.TRANSACTION, "channel_order_no": proxyPayRespVO.Data.ChannelOrderNo}).Error; err != nil {
			logx.WithContext(l.ctx).Errorf("更新下发子单状态之败:%s", err.Error())
		}
	}

	return nil
}
