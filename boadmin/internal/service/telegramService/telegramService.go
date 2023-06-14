package telegramService

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

func SendMessage(ctx context.Context, svcCtx *svc.ServiceContext, chatId, message string) (sendResponse types.TelegramSendResponse, err error) {
	url := fmt.Sprintf("%s:20003/telegram/send", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	req := types.TelegramSendRequest{
		ChatID:  chatId,
		Message: message,
	}
	resp, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("telegram sendMessage error:%s", err.Error())
		return sendResponse, errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if resp.Status() != 200 {
		logx.WithContext(ctx).Infof("telegram sendMessage error:Status: %d  Body: %s", resp.Status(), string(resp.Body()))
		return sendResponse, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("Error HTTP Status: %d", resp.Status()))
	}

	sendResponse = types.TelegramSendResponse{}
	// 返回body 轉 struct
	if err = resp.DecodeJSON(&sendResponse); err != nil {
		return sendResponse, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if sendResponse.Code != "0" {
		logx.WithContext(ctx).Errorf("telegram sendMessage error:%s", sendResponse.Message)
		return sendResponse, errorz.New(response.GENERAL_ERROR)
	}
	return
}

func EditMessage(ctx context.Context, svcCtx *svc.ServiceContext, chatId, messageId, message string) (editResponse types.TelegramEditResponse, err error) {
	url := fmt.Sprintf("%s:20003/telegram/edit", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	req := types.TelegramEditRequest{
		ChatID:    chatId,
		MessageID: messageId,
		Message:   message,
	}
	resp, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("telegram editMessage error:%s", err.Error())
		return editResponse, errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if resp.Status() != 200 {
		logx.WithContext(ctx).Infof("telegram editMessage error:Status: %d  Body: %s", resp.Status(), string(resp.Body()))
		return editResponse, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("Error HTTP Status: %d", resp.Status()))
	}

	editResponse = types.TelegramEditResponse{}
	// 返回body 轉 struct
	if err = resp.DecodeJSON(&editResponse); err != nil {
		return editResponse, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if editResponse.Code != "0" {
		logx.WithContext(ctx).Errorf("telegram editMessage error:%s", editResponse.Message)
		return editResponse, errorz.New(response.GENERAL_ERROR)
	}
	return
}

func DeleteMessage(ctx context.Context, svcCtx *svc.ServiceContext, chatId, messageId string) (deleteResponse types.TelegramDeleteResponse, err error) {
	url := fmt.Sprintf("%s:20003/telegram/delete", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	req := types.TelegramDeleteRequest{
		ChatID:    chatId,
		MessageID: messageId,
	}
	resp, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(req)
	if err != nil {
		logx.WithContext(ctx).Errorf("telegram deleteMessage error:%s", err.Error())
		return deleteResponse, errorz.New(response.DATABASE_FAILURE, err.Error())
	} else if resp.Status() != 200 {
		logx.WithContext(ctx).Infof("telegram deleteMessage error:Status: %d  Body: %s", resp.Status(), string(resp.Body()))
		return deleteResponse, errorz.New(response.INVALID_STATUS_CODE, fmt.Sprintf("Error HTTP Status: %d", resp.Status()))
	}

	deleteResponse = types.TelegramDeleteResponse{}
	// 返回body 轉 struct
	if err = resp.DecodeJSON(&deleteResponse); err != nil {
		return deleteResponse, errorz.New(response.GENERAL_EXCEPTION, err.Error())
	}

	if deleteResponse.Code != "0" {
		logx.WithContext(ctx).Errorf("telegram deleteMessage error:%s", deleteResponse.Message)
		return deleteResponse, errorz.New(response.GENERAL_ERROR)
	}
	return
}
