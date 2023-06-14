package orderrecord

import (
	"com.copo/bo_service/boadmin/internal/logic/orderrecord"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func ReceiptRecordExcelHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReceiptRecordQueryAllRequestX

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseJsonBody(r, &req); err != nil {
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

		req.Language = r.Header.Get("Accept-Language")

		l := orderrecord.NewReceiptRecordExcelLogic(r.Context(), ctx)
		err := l.ReceiptRecordExcel(&req)

		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			//fileName := "ReceiptReport" + time.Now().Format("20060102150405") + ".xlsx"
			//w.Header().Set("Content-Type", "application/octet-stream")
			//w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
			//w.Header().Set("Content-Transfer-Encoding", "binary")
			//w.Header().Set("Expires", "0")
			//userLogService.CreateUserLog(r, req, ctx)
			//xlsx.Write(w)
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
