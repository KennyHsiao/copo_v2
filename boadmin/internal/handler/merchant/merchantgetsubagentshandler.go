package merchant

import (
	"com.copo/bo_service/boadmin/internal/logic/merchant"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/common/response"
	"net/http"
)

func MerchantGetSubAgentsHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := merchant.NewMerchantGetSubAgentsLogic(r.Context(), ctx)
		resp, err := l.MerchantGetSubAgents()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, nil, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
