package withdraworder

import (
	"com.copo/bo_service/boadmin/internal/logic/withdraworder"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/common/response"
	"net/http"
)

func WithdrawApiTestCallBackHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := withdraworder.NewWithdrawApiTestCallBackLogic(r.Context(), ctx)
		resp, err := l.WithdrawApiTestCallBack()
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			w.Write([]byte(resp))
		}
	}
}
