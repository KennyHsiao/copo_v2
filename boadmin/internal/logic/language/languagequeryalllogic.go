package language

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type LanguageQueryAllLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLanguageQueryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) LanguageQueryAllLogic {
	return LanguageQueryAllLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LanguageQueryAllLogic) LanguageQueryAll(req types.LanguageQueryAllRequest) (resp *types.LanguageQueryAllResponse, err error) {
	var languages []types.Language
	pageNum := req.PageNum - 1
	offset := pageNum * req.PageSize
	var count int64
	//var terms []string

	db := l.svcCtx.MyDB.Table("bs_lang")

	if len(req.Name) > 0 {
		//terms = append(terms, fmt.Sprintf("name like '%%%s%%'", req.Name))
		db = db.Where("a.mission_name LIKE ?", "%"+req.Name+"%")
	}
	//term := strings.Join(terms, " AND ")
	db.Count(&count)
	err = db.Offset(offset).Limit(req.PageSize).Find(&languages).Error

	resp = &types.LanguageQueryAllResponse{
		List:     languages,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		RowCount: count,
	}
	return

}
