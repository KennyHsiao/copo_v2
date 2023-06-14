package order

import (
	"com.copo/bo_service/boadmin/internal/logic/order"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/common/response"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func WithdrawVerifyWayHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		defer span.End()

		l := order.NewWithdrawVerifyWayLogic(r.Context(), ctx)
		resp, err := l.WithdrawVerifyWay()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
