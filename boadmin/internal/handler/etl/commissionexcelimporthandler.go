package etl

import (
	"com.copo/bo_service/boadmin/internal/logic/etl"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func CommissionExcelImportHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadExcelRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		//if err := httpx.ParseForm(r, &req); err != nil {
		//	response.Json(w, r, response.FAIL, nil, err)
		//	return
		//}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			return
		}

		file, _, errFile := r.FormFile("uploadFile")
		if errFile != nil {
			response.Json(w, r, response.FAIL, nil, errFile)
			return
		}
		req.UploadFile = file

		if requestBytes, err := json.Marshal(req); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
		}

		l := etl.NewCommissionExcelImportLogic(r.Context(), ctx)
		err := l.CommissionExcelImport(&req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
