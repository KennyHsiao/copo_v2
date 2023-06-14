package systemrate

import (
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/common/response"
	"net/http"

	"com.copo/bo_service/boadmin/internal/logic/systemrate"
	"com.copo/bo_service/boadmin/internal/svc"
)

func SystemChannelBalanceLimitQueryHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := systemrate.NewSystemChannelBalanceLimitQueryLogic(r.Context(), ctx)
		resp, err := l.SystemChannelBalanceLimitQuery()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
