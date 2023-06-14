package commissionWithdrawOrder

import (
	"com.copo/bo_service/boadmin/internal/logic/commissionWithdrawOrder"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"mime/multipart"
	"net/http"
	"strings"
)

func CommissionWithdrawOrderCreateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CommissionWithdrawOrderCreateRequest

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			httpx.Error(w, err)
			return
		}

		files := []*multipart.FileHeader{}

		for key, formFiles := range r.MultipartForm.File {
			if strings.HasPrefix(key, "uploadFile") {
				files = formFiles
			}

		}

		if err := httpx.ParseForm(r, &req); err != nil {
			response.Json(w, r, response.FAIL, nil, err)
			return
		}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := commissionWithdrawOrder.NewCommissionWithdrawOrderCreateLogic(r.Context(), ctx)
		err := l.CommissionWithdrawOrderCreate(&req, files)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
