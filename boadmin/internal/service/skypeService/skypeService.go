package skypeService

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
)

func SendMessage(ctx context.Context, svcCtx *svc.ServiceContext, announcementMerchantId, chatId, message string) (sendResponse types.SkypeSendResponse, err error) {
	url := fmt.Sprintf("%s:20003/skype/send", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	req := types.SkypeSendRequest{
		ID:     announcementMerchantId,
		ChatId: chatId,
		Text:   message,
	}
	resp, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("skype sendMessage error:%s", err.Error())
		return sendResponse, errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if resp.Status() != 200 {
		logx.WithContext(ctx).Infof("skype sendMessage error:Status: %d  Body: %s", resp.Status(), string(resp.Body()))
		return sendResponse, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("Error HTTP Status: %d", resp.Status()))
	}

	sendResponse = types.SkypeSendResponse{}
	// 返回body 轉 struct
	if err = resp.DecodeJSON(&sendResponse); err != nil {
		return sendResponse, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if sendResponse.Code != "0" {
		logx.WithContext(ctx).Errorf("skype sendMessage error:%s", sendResponse.Message)
		return sendResponse, errorz.New(response.GENERAL_ERROR)
	}
	return
}
