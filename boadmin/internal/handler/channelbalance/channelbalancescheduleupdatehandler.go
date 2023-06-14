package channelbalance

import (
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/common/response"
	"net/http"

	"com.copo/bo_service/boadmin/internal/logic/channelbalance"
	"com.copo/bo_service/boadmin/internal/svc"
)

func ChannelBalanceScheduleUpdateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := channelbalance.NewChannelBalanceScheduleUpdateLogic(r.Context(), ctx)
		err := l.ChannelBalanceScheduleUpdate()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
