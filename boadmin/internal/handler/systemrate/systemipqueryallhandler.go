package systemrate

import (
	"com.copo/bo_service/boadmin/internal/logic/systemrate"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/common/response"
	"net/http"
)

func SystemIpQueryAllHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := systemrate.NewSystemIpQueryAllLogic(r.Context(), ctx)
		resp, err := l.SystemIpQueryAll()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
