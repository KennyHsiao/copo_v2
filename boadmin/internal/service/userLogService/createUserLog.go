package userLogService

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"github.com/thinkeridea/go-extend/exnet"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"strings"
)

func CreateUserLog(r *http.Request, req interface{}, svcCtx *svc.ServiceContext) (err error) {

	jwtMerchantCode := ""
	jwtAccount := ""
	jwtName := ""
	if v := r.Context().Value("merchantCode"); v != nil {
		jwtMerchantCode = r.Context().Value("merchantCode").(string)
	}
	if v := r.Context().Value("account"); v != nil {
		jwtAccount = r.Context().Value("account").(string)
	}
	if v := r.Context().Value("name"); v != nil {
		jwtName = r.Context().Value("name").(string)
	}

	path := r.URL.Path
	ctx := r.Context()
	requestParam := ""
	reqParamMap := map[string]interface{}{}
	// 取得API 模版
	isNeedCreate, template := getUserLogTemplate(ctx, svcCtx, path, jwtMerchantCode)
	if !isNeedCreate {
		return
	}

	if req != nil {
		bodyBytes, err := json.Marshal(req)
		if err != nil {
			logx.WithContext(ctx).Errorf("建立user log 失敗: %s", err.Error())
			return err
		}
		requestParam = string(bodyBytes)
		json.Unmarshal(bodyBytes, &reqParamMap)

	}

	reqParamMap["jwtName"] = jwtName
	reqParamMap["jwtMerchantCode"] = jwtMerchantCode
	reqParamMap["jwtAccount"] = jwtAccount
	// 登入时没有 jwtAccount
	if path == "/api/v1/auth/login" {
		jwtAccount = fmt.Sprint(reqParamMap["account"])
	}

	userLog := types.UserLog{
		AccountName:  jwtAccount,
		Type:         template.Type,
		RequestParam: requestParam,
		RequestApi:   template.ApiName,
		RequestUrl:   path,
		Method:       r.Method,
		RequestUnit:  template.ApiUnit,
		UserAgent:    r.UserAgent(),
		Ip:           exnet.ClientIP(r),
		Operating:    formatOperating(template, reqParamMap),
	}

	if err = svcCtx.MyDB.Table("au_user_log").Create(&types.UserLogX{
		UserLog: userLog,
	}).Error; err != nil {
		logx.WithContext(ctx).Errorf("建立user log 失敗: %s", err.Error())
		return err
	}

	return
}

// 使用請求路徑找到相對應的模板
func getUserLogTemplate(ctx context.Context, svcCtx *svc.ServiceContext, path, jwtMerchantCode string) (isNeedCreate bool, template types.UserLogTemplateX) {
	var templates []types.UserLogTemplateX
	if err := svcCtx.MyDB.Table("au_user_log_template").
		Where("path = ?", path).Order("user_type desc").
		Find(&templates).Error; err != nil {
		logx.WithContext(ctx).Errorf("建立user log 失敗: %s", err.Error())
		return
	}

	// 某些API 模板分為 商戶模板 管理員模板
	if len(templates) == 0 {
		return
	} else if len(templates) >= 2 {
		if len(jwtMerchantCode) == 0 {
			// 利用SQL排序,管理員index = 0
			template = templates[0]
		} else {
			template = templates[1]
		}
	} else {
		// 只有一個模板則直接使用
		template = templates[0]
	}
	isNeedCreate = true
	return
}

func formatOperating(template types.UserLogTemplateX, paramMap map[string]interface{}) string {
	for k, v := range paramMap {
		key := "%" + k + "%"
		template.MsgTemplate = strings.Replace(template.MsgTemplate, key, fmt.Sprint(v), -1)
	}
	return template.MsgTemplate
}
