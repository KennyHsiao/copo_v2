package withdraworder

import (
	"com.copo/bo_service/boadmin/internal/logic/withdraworder"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func WithdrawApiOrderHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req1 types.WithdrawApiOrderRequest

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req1); err != nil {
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}

		var req types.WithdrawApiOrderRequestX
		req.WithdrawApiOrderRequest = req1

		myIP := exnet.ClientIP(r)
		req.MyIp = myIP

		if err := utils.MyValidator.Struct(req); err != nil {
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := withdraworder.NewWithdrawApiOrderLogic(r.Context(), ctx)
		resp, err := l.WithdrawApiOrder(req)
		if err != nil {
			response.ApiErrorJson(w, r, err.Error(), err)
		} else {
			response.ApiJson(w, r, resp)
		}
	}
}
