package download

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/gormx"
	"com.copo/bo_service/common/response"
	"com.copo/bo_service/common/utils"
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadCenterQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadCenterQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) DownloadCenterQueryAllLogic {
	return DownloadCenterQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadCenterQueryAllLogic) DownloadCenterQueryAll(req types.DownloadCenterQueryAllRequest) (resp *types.DownloadCenterQueryAllResponseX, err error) {
	account := l.ctx.Value("account").(string)
	var downDatas []types.DownloadCenterX
	var count int64
	var terms []string

	db := l.svcCtx.MyDB.Table("rp_down_report a").
		Joins("LEFT JOIN au_users b on a.user_id = b.id")

	if len(req.UserId) > 0 {
		terms = append(terms, fmt.Sprintf("b.account = '%s'", req.UserId))
		db = db.Where("b.account = ?", req.UserId)
	} else {
		//terms = append(terms, fmt.Sprintf("b.account = '%s'", account))
		db = db.Where("b.account = ?", account)
	}
	if len(req.Type) > 0 {
		//terms = append(terms, fmt.Sprintf("a.type = '%s'", req.Type))
		db = db.Where("a.type = ?", req.Type)
	}
	if len(req.MissionName) > 0 {
		//terms = append(terms, fmt.Sprintf("a.mission_name like '%%%s%%'", req.MissionName))
		db = db.Where("a.mission_name LIKE ?", "%"+req.MissionName+"%")
	}
	if len(req.Status) > 0 {
		//terms = append(terms, fmt.Sprintf("a.status = '%s'", req.Status))
		db = db.Where("a.status = ?", req.Status)
	}
	if len(req.StartAt) > 0 {
		terms = append(terms, fmt.Sprintf(" a.created_at >= '%s'", req.StartAt))
		db = db.Where("a.created_at >= ?", req.StartAt)
	}
	if len(req.EndAt) > 0 {
		endAt := utils.ParseTimeAddOneSecond(req.EndAt)
		//terms = append(terms, fmt.Sprintf(" a.created_at < '%s'", endAt))
		db = db.Where("a.created_at < ?", endAt)
	}

	//term := strings.Join(terms, " AND ")

	selectX := "a.id," +
		"a.user_id," +
		"a.type," +
		"a.file_name," +
		"a.`status`," +
		"a.created_at," +
		"a.file_path," +
		"a.mission_name," +
		"b.`account` as user_name"
	db.Count(&count)
	err = db.
		Scopes(gormx.Paginate(req)).
		Select(selectX).
		Order("a.created_at desc").
		Find(&downDatas).Error
	if err != nil {
		return nil, errorz.New(response.DATABASE_FAILURE, err.Error())
	}

	resp = &types.DownloadCenterQueryAllResponseX{
		List:     downDatas,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}

	return
}
