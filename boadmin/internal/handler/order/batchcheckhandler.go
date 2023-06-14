package order

import (
	"com.copo/bo_service/boadmin/internal/logic/order"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"github.com/zeromicro/go-zero/rest/httpx"

	"net/http"
)

func BatchCheckHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BatchCheckOrderRequest

		if err := httpx.ParseJsonBody(r, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		l := order.NewBatchCheckLogic(r.Context(), ctx)
		resp, err := l.BatchCheck(req)
		if err != nil {
			response.Json(w, r, err.Error(), resp, err)
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
