package payorder

import (
	"com.copo/bo_service/boadmin/internal/logic/payorder"
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

func PayQueryBalanceHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PayQueryBalanceRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.ApiErrorJson(w, r, response.API_INVALID_PARAMETER, err)
			return
		}

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

		myIP := exnet.ClientIP(r)
		req.MyIp = myIP

		l := payorder.NewPayQueryBalanceLogic(r.Context(), ctx)
		resp, err := l.PayQueryBalance(req)
		if err != nil {
			response.ApiErrorJson(w, r, err.Error(), err)
		} else {
			response.ApiJson(w, r, resp)
		}
	}
}
