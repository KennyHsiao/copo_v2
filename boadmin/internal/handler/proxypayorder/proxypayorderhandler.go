package proxypayorder

import (
	"com.copo/bo_service/boadmin/internal/logic/proxypayorder"
	transactionLogService "com.copo/bo_service/boadmin/internal/service/transactionLog"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/constants"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"fmt"
	"github.com/gioco-play/easy-i18n/i18n"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func ProxyPayOrderHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ProxyPayRequestX
		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			logx.Error("Parse Error:", err.Error())
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			logx.Error("Validatation Error: ", err.Error())
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}
		req.Ip = exnet.ClientIP(r)

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := proxypayorder.NewProxyPayOrderLogic(r.Context(), ctx)
		resp, err := l.ProxyPayOrder(&req)
		if err != nil {
			var msg string
			if v, ok := err.(*errorz.Err); ok && v.GetMessage() != "" {
				msg = v.GetMessage()
			} else {
				msg = i18n.Sprintf("%s", err.Error())
			}
			// 写入交易日志
			contentStrut := struct {
				ErrorCode string
				ErrorMsg  string
			}{
				ErrorCode: fmt.Sprintf("%s", err.Error()),
				ErrorMsg:  msg,
			}

			if errLog := transactionLogService.CreateTransactionLog(ctx.MyDB, &types.TransactionLogData{
				MerchantCode:    req.MerchantId,
				MerchantOrderNo: req.OrderNo,
				//OrderNo:         req,
				LogType:   constants.ERROR_REPLIED_TO_MERCHANT,
				LogSource: constants.API_DF,
				Content:   contentStrut,
				TraceId:   trace.SpanContextFromContext(r.Context()).TraceID().String(),
			}); errLog != nil {
				logx.WithContext(r.Context()).Errorf("写入交易日志错误:%s", errLog)
			}

			response.ApiErrorJson(w, r, err.Error(), err)
		} else {
			response.ApiJson(w, r, resp)
		}
	}
}
