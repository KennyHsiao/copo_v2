package order

import (
	"com.copo/bo_service/boadmin/internal/logic/order"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strconv"
)

func InternalChargeOrderCreateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req1 types.OrderInternalCreateRequest
		var req types.OrderInternalCreate

		errFile := r.ParseMultipartForm(32 << 20)
		if errFile != nil {
			response.Json(w, r, response.FAIL, nil, errFile)
			return
		}
		formData := r.MultipartForm.File
		var orderX types.OrderX
		orderX.MerchantBankAccount = r.MultipartForm.Value["merchantBankAccount"][0]
		orderX.MerchantBankNo = r.MultipartForm.Value["merchantBankNo"][0]
		orderX.MerchantAccountName = r.MultipartForm.Value["merchantAccountName"][0]
		orderX.ChannelBankAccount = r.MultipartForm.Value["channelBankAccount"][0]
		orderX.ChannelBankNo = r.MultipartForm.Value["channelBankNo"][0]
		orderX.ChannelAccountName = r.MultipartForm.Value["channelAccountName"][0]
		orderX.OrderAmount, _ = strconv.ParseFloat(r.MultipartForm.Value["orderAmount"][0], 64)
		orderX.CurrencyCode = r.MultipartForm.Value["currencyCode"][0]
		orderX.MerchantBankName = r.MultipartForm.Value["merchantBankName"][0]
		orderX.ChannelBankName = r.MultipartForm.Value["channelBankName"][0]
		req.OrderX = orderX
		req.FormData = formData

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseForm(r, &req1); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := order.NewInternalChargeOrderCreateLogic(r.Context(), ctx)
		err := l.InternalChargeOrderCreate(req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
