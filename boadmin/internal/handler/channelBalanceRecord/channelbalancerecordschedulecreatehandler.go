package channelBalanceRecord

import (
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/common/response"
	"net/http"

	"com.copo/bo_service/boadmin/internal/logic/channelBalanceRecord"
	"com.copo/bo_service/boadmin/internal/svc"
)

func ChannelBalanceRecordScheduleCreateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := channelBalanceRecord.NewChannelBalanceRecordScheduleCreateLogic(r.Context(), ctx)
		err := l.ChannelBalanceRecordScheduleCreate()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
