package admin_user

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/response"
	"context"

	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) UserQueryLogic {
	return UserQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserQueryLogic) UserQuery(req types.UserQueryRequest) (resp *types.UserQueryResponse, err error) {
	merchantCode := l.ctx.Value("merchantCode").(string)
	ux := model.NewUser(l.svcCtx.MyDB)
	resp, err = ux.GetUser(req.ID)

	// 若登入者是商戶 要限制他能查詢的帳號
	if merchantCode != "" {
		if len(resp.Merchants) == 0 { // 這表示此為管理員帳號
			return nil, errorz.New(response.DATABASE_FAILURE, "")
		} else if resp.Merchants[0].Code != merchantCode &&
			resp.Merchants[0].AgentParentCode != merchantCode { // 這表示非此商戶的子帳號帳號 或 子代理帳號 不可訪問
			return nil, errorz.New(response.DATABASE_FAILURE, "")
		}
	}

	return
}
