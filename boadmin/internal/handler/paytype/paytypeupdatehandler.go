package paytype

import (
	"com.copo/bo_service/boadmin/internal/logic/paytype"
	"com.copo/bo_service/boadmin/internal/service/userLogService"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strconv"
)

func PayTypeUpdateHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req1 types.PayTypeUpdateRequest
		var req types.PayTypeUpdate

		r.ParseMultipartForm(32 << 20)
		if len(r.MultipartForm.File) > 0 {
			file, header, errFile := r.FormFile("uploadFile")
			if errFile != nil {
				logx.Error(errFile.Error())
				response.Json(w, r, response.FAIL, nil, errFile)
				return
			}
			req.UploadFile = file
			req.UploadHeader = header
		}
		req.ID, _ = strconv.ParseInt(r.MultipartForm.Value["id"][0], 10, 64)
		req.Code = r.MultipartForm.Value["code"][0]
		req.Name = r.MultipartForm.Value["name"][0]
		req.Currency = r.MultipartForm.Value["currency"][0]
		req.NameI18n.En = r.MultipartForm.Value["nameEN"][0]
		req.NameI18n.Zh = r.MultipartForm.Value["name"][0]
		//req.SortNum = r.MultipartForm.Value["sortNum"][0]

		span := trace.SpanFromContext(r.Context())
		defer span.End()

		if err := httpx.ParseForm(r, &req1); err != nil {
			logx.Error(err.Error())
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

		l := paytype.NewPayTypeUpdateLogic(r.Context(), ctx)
		err := l.PayTypeUpdate(req)
		if err != nil {
			response.Json(w, r, err.Error(), nil, err)
		} else {
			userLogService.CreateUserLog(r, req, ctx)
			response.Json(w, r, response.SUCCESS, nil, err)
		}
	}
}
