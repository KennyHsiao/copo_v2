package callNoticeUrlService

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"fmt"
	"github.com/gioco-play/gozzle"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

func CallNoticeUrlForChannelBalanceFail(ctx context.Context, svcCtx *svc.ServiceContext, failList []string) error {
	msg := append(failList, "余额更新失败!!请手动更新馀额。")
	notifyMsg := types.TelegramNotifyRequest{
		Message: strings.Join(msg, ", \n"),
	}

	url := fmt.Sprintf("%s:20003/telegram/notify_balance", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	if _, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(notifyMsg); err != nil {
		logx.WithContext(ctx).Errorf("馀额报警通知失敗:%s", err.Error())
	}

	return nil
}

func CallNoticeUrlForChannelBalanceSuccess(ctx context.Context, svcCtx *svc.ServiceContext, msg types.TelegramNotifyRequest) error {
	url := fmt.Sprintf("%s:20003/telegram/notify_balance", svcCtx.Config.Server)
	span := trace.SpanFromContext(ctx)
	if _, err := gozzle.Post(url).Timeout(25).Trace(span).JSON(msg); err != nil {
		logx.WithContext(ctx).Errorf("馀额报警通知失敗:%s", err.Error())
	}

	return nil
}
