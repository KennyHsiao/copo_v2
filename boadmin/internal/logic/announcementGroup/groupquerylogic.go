package announcementGroup

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GroupQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGroupQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) GroupQueryLogic {
	return GroupQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GroupQueryLogic) GroupQuery(req types.GroupQueryRequest) (resp *types.GroupQueryResponseX, err error) {
	var groupQueries []types.GroupQueryX
	var count int64

	db := l.svcCtx.MyDB
	if len(req.MerchantCode) > 0 {
		db = db.Where("a.code = ?", req.MerchantCode)
	}

	if len(req.CommunicationSoftware) > 0 {
		if req.CommunicationSoftware == "1" {
			db = db.Where("a.contact->> '$.communicationSoftware' = ?", "telegram")
		} else if req.CommunicationSoftware == "2" {
			db = db.Where("a.contact->> '$.communicationSoftware' = ?", "skype")
		}
	}

	if len(req.GroupName) > 0 {
		db = db.Where("a.contact->> '$.groupName' like ?", "%"+req.GroupName+"%")
	}

	if len(req.GroupId) > 0 {
		db = db.Where("a.contact->> '$.groupId' = ?", req.GroupId)
	} else {
		db = db.Where("a.contact->> '$.groupId' != \"\"")
	}

	selectX := "a.contact ->> \"$.groupId\" as group_id," +
		"a.contact ->> \"$.groupName\" as group_name," +
		"a.contact ->> \"$.communicationSoftware\" as communication_software," +
		"a.contact ->> \"$.communicationNickname\" as communication_nickname," +
		"a.code as merchant_code," +
		"created_at," +
		"updated_at"

	if err = db.Table("mc_merchants a").Count(&count).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	if err = db.Table("mc_merchants as a").
		Scopes(gormx.Paginate(req)).
		Select(selectX).Find(&groupQueries).Error; err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.GroupQueryResponseX{
		List:     groupQueries,
		PageSize: req.PageSize,
		PageNum:  req.PageNum,
		RowCount: count,
	}
	return
}
