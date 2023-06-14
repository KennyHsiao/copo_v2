package etl

import (
	"com.copo/bo_service/boadmin/internal/logic/etl"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/common/response"
	"net/http"
)

func ChannelEtlHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := etl.NewChannelEtlLogic(r.Context(), ctx)
		err := l.ChannelEtl()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
