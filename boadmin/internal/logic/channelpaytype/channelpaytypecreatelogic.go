package channelpaytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"sort"
	"strings"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
)

type ChannelPayTypeCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChannelPayTypeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) ChannelPayTypeCreateLogic {
	return ChannelPayTypeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChannelPayTypeCreateLogic) ChannelPayTypeCreate(req types.ChannelPayTypeCreateRequest) error {
	return l.svcCtx.MyDB.Transaction(func(tx *gorm.DB) error {
		_ = &types.ChannelPayTypeCreate{
			ChannelPayTypeCreateRequest: req,
		}
		//判斷Code 是否重複
		err := model.NewChannelPayType(l.svcCtx.MyDB).CheckChannelPayTypeDuplicated(req)

		if err != nil {
			logx.Error("渠道配置支付方式錯誤: ", err.Error())
			return err
		}

		//排序固定金額
		if req.FixedAmount != "" {
			amountArr := strings.Split(req.FixedAmount, ",")
			sort.Strings(amountArr)
			fixedAmount := strings.Join(amountArr, ",")
			req.FixedAmount = fixedAmount
		}

		return nil
	})
}
