package order

import (
	"com.copo/bo_service/boadmin/internal/logic/order"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

func OrderImageUploadHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadImageRequestX
		r.ParseMultipartForm(32 << 20)
		files := r.MultipartForm.File
		file, header, errFile := r.FormFile("uploadFile")
		if errFile != nil {
			response.Json(w, r, response.FAIL, nil, errFile)
			return
		}
		req.UploadFile = file
		req.UploadHeader = header
		req.Files = files

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		//if err := httpx.ParseJsonBody(r, &req); err != nil {
		//	response.Json(w, r, response.FAIL, nil, err)
		//	return
		//}

		if err := utils.MyValidator.Struct(req); err != nil {
			response.Json(w, r, response.INVALID_PARAMETER, nil, err)
			logx.Info("参数错误: ", err.Error())
			return
		}

		if requestBytes, err := json.Marshal(r); err == nil {
			span.SetAttributes(attribute.KeyValue{
				Key:   "request",
				Value: attribute.StringValue(string(requestBytes)),
			})
			logx.Info("参数错误: ", err.Error())
		}

		l := order.NewOrderImageUploadLogic(r.Context(), ctx)
		resp, err := l.OrderImageUpload(req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
			logx.Info("上传错误: ", err.Error())
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, resp, err)
		}
	}
}
