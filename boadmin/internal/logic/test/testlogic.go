package test

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"go.opentelemetry.io/otel/trace"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) TestLogic {
	return TestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TestLogic) Test(req *types.TestRequest) (resp *types.TestResponse, err error) {

	span := trace.SpanFromContext(l.ctx)
	res, _ := gozzle.Get("http://127.0.0.1:8080/api/v1/test/healthCheck").
		Timeout(20).Trace(span).Do()

	channelResp := struct {
		Success bool        `json:"success"`
		Message string      `json:"message"`
		Errors  interface{} `json:"errors"`
	}{}

	// 返回body 轉 struct
	if err = res.DecodeJSON(&channelResp); err != nil {
		return nil, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if channelResp.Errors != nil {
		logx.WithContext(l.ctx).Info(fmt.Sprintf("%s", channelResp.Errors))
	} else {
		logx.WithContext(l.ctx).Info("nil")
	}

	return
}
