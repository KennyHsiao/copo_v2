package callNoticeUrlService

import (
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/apimodel/vo"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"time"
)

func CallNoticeUrlForZF(ctx context.Context, db *gorm.DB, notifyUrl string, payCallBackVO vo.PayCallBackVO) error {

	// 写入交易日志
	createTransactionLog(ctx, db, payCallBackVO, constants.CALLBACK_TO_MERCHANT, payCallBackVO)

	span := trace.SpanFromContext(ctx)
	res, errx := gozzle.Post(notifyUrl).Timeout(15).Trace(span).JSON(payCallBackVO)
	if errx != nil {
		logx.WithContext(ctx).Errorf("call NotifyUrl error:%s", errx.Error())
		// 写入交易日志
		createTransactionLog(ctx, db, payCallBackVO, constants.RESPONSE_FROM_MERCHANT, fmt.Sprintf("call NotifyUrl error:%s", errx.Error()))
		return errorz.New(response.MERCHANT_CALLBACK_ERROR, errx.Error())
	} else if res.Status() != 200 {
		logx.WithContext(ctx).Errorf("call NotifyUrl httpStatus:%d, res:%s", res.Status(), string(res.Body()[:]))
		// 写入交易日志
		createTransactionLog(ctx, db, payCallBackVO, constants.RESPONSE_FROM_MERCHANT, fmt.Sprintf("call NotifyUrl httpStatus:%d, res:%s", res.Status(), string(res.Body()[:])))
		return errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("call NotifyUrl httpStatus:%d", res.Status()))
	}
	respString := string(res.Body()[:])

	// 写入交易日志
	createTransactionLog(ctx, db, payCallBackVO, constants.RESPONSE_FROM_MERCHANT, respString)

	// 商戶回覆確認
	logx.WithContext(ctx).Errorf("UI支付回调商户, 商户回传:%s", respString)
	if respString == "success" {
		if err4 := db.Table("tx_orders").
			Where("order_no = ?", payCallBackVO.PayOrderId).
			Updates(map[string]interface{}{"is_merchant_callback": "1", "merchant_call_back_at": time.Now()}).Error; err4 != nil {
			logx.WithContext(ctx).Errorf("回調成功,但更改回調狀態失敗")
			return errorz.New(response.MERCHANT_CALLBACK_ERROR, "回調成功,但更改回調狀態失敗")
		}
	} else {
		logx.WithContext(ctx).Errorf("商户回复错误:%s", respString)
		return errorz.New(response.MERCHANT_CALLBACK_ERROR)
	}
	return nil
}

func createTransactionLog(ctx context.Context, db *gorm.DB, payCallBackVO vo.PayCallBackVO, logType string, content interface{}) {
	// 写入交易日志
	var errLog error
	if errLog = transactionLogService.CreateTransactionLog(db, &types.TransactionLogData{
		MerchantCode:    payCallBackVO.MerchantId,
		MerchantOrderNo: payCallBackVO.OrderNo,
		OrderNo:         payCallBackVO.PayOrderId,
		LogType:         logType,
		LogSource:       constants.API_ZF,
		Content:         content,
		TraceId:         trace.SpanContextFromContext(ctx).TraceID().String(),
	}); errLog != nil {
		logx.WithContext(ctx).Errorf("写入交易日志错误:%s", errLog)
	}
}
