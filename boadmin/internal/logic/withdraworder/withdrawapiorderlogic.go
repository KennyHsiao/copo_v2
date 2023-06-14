package withdraworder

import (
	"com.copo/bo_service/boadmin/internal/service/merchantsService"
	ordersService "com.copo/bo_service/boadmin/internal/service/orders"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"errors"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
)

type WithdrawApiOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawApiOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) WithdrawApiOrderLogic {
	return WithdrawApiOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WithdrawApiOrderLogic) WithdrawApiOrder(req types.WithdrawApiOrderRequestX) (resp *types.WithdrawApiOrderResponse, err error) {
	db := l.svcCtx.MyDB
	var orderWithdrawCreateResp *types.OrderWithdrawCreateResponse
	var newOrder types.OrderX
	var merchant types.Merchant
	// 檢查白名單
	if err = db.Table("mc_merchants").Where("code = ?", req.MerchantId).Take(&merchant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorz.New(response.DATA_NOT_FOUND, err.Error())
		} else {
			return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
		}
	}

	if isWhite := merchantsService.IPChecker(req.MyIp, merchant.ApiIP); !isWhite {
		return nil, errorz.New(response.API_IP_DENIED, "IP: "+req.MyIp)
	}

	// 驗簽檢查
	if isSameSign := utils.VerifySign(req.Sign, req.WithdrawApiOrderRequest, merchant.ScrectKey); !isSameSign {
		return nil, errorz.New(response.INVALID_SIGN)
	}

	orderAmount, errParse := strconv.ParseFloat(req.OrderAmount, 64)
	if errParse != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, errParse.Error())
	}

	//确认是否重复订单
	var isExist bool
	if err = db.Table("tx_orders").
		Select("count(*) > 0 ").
		Where("merchant_code = ? AND merchant_order_no = ?", req.MerchantId, req.OrderNo).
		Find(&isExist).Error; err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION)
	}
	if isExist {
		return nil, errorz.New(response.REPEAT_ORDER_NO)
	}

	var withdrawOrders []types.OrderWithdrawCreateRequestX
	var withdrawOrder types.OrderWithdrawCreateRequestX
	withdrawOrder.Type = "XF"
	withdrawOrder.MerchantAccountName = req.WithdrawName
	withdrawOrder.MerchantBankName = req.BankName
	withdrawOrder.MerchantBankProvince = req.BankProvince
	withdrawOrder.MerchantBankCity = req.BankCity
	withdrawOrder.MerchantBankAccount = req.AccountNo
	withdrawOrder.CurrencyCode = req.Currency
	withdrawOrder.OrderAmount = orderAmount
	withdrawOrder.Source = constants.API
	withdrawOrder.MerchantCode = req.MerchantId
	withdrawOrder.UserAccount = req.MerchantId
	withdrawOrder.NotifyUrl = req.NotifyUrl
	withdrawOrder.MerchantOrderNo = req.OrderNo
	withdrawOrder.PageUrl = req.PageUrl

	withdrawOrders = append(withdrawOrders, withdrawOrder)

	orderWithdrawCreateResp, err = ordersService.WithdrawOrderCreate(db, withdrawOrders, l.ctx, l.svcCtx)
	if err != nil {
		logx.Error("err: ", err.Error())
		//tx.Rollback()
		return nil, err
	}
	newOrder = orderWithdrawCreateResp.OrderX
	//tx.Commit()

	newData := make(map[string]string)
	newData["withdrawOrderNo"] = newOrder.OrderNo
	newSign := utils.SortAndSign(newData, merchant.ScrectKey)

	respData := types.RespData{
		WithdrawOrderNo: newOrder.OrderNo,
		Sign:            newSign,
	}
	resp = &types.WithdrawApiOrderResponse{
		RespCode: response.API_SUCCESS,
		RespMsg:  i18n.Sprintf(response.API_SUCCESS), //固定回商戶成功
		RespData: respData,
	}

	return
}
