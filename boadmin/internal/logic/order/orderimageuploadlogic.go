package order

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"context"
	"io"
	"os"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderImageUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderImageUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) OrderImageUploadLogic {
	return OrderImageUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderImageUploadLogic) OrderImageUpload(req types.UploadImageRequestX) (resp []string, err error) {
	var splitStrs []string
	//var header = req.UploadHeader
	files := req.Files
	//aaa := files["uploadFile"]
	for _, ff := range files {
		ext := strings.ToLower(path.Ext(ff[0].Filename))
		if ext != ".jpg" && ext != ".png" {
			return nil, errorz.New(response.FILE_TYPE_NOT_JPG_ERROR)
		}
		file, err := ff[0].Open()
		if err != nil {

		}
		defer file.Close()
		var terms []string
		randStr := random.GetRandomString(10, random.ALL, random.MIX)
		terms = append(append(terms, randStr), ext)
		newFileName := strings.Join(terms, "")
		f, errOpenFile := os.OpenFile("./public/uploads/internalcharges/"+newFileName, os.O_WRONLY|os.O_CREATE, 0777)
		if errOpenFile != nil {
			return nil, errorz.New(response.FAIL, err.Error())
		}
		defer f.Close()
		//把.去掉
		splitStr := strings.Split(f.Name(), ".")
		io.Copy(f, file)
		splitStrs = append(splitStrs, splitStr[1])
	}
	return splitStrs, nil
}
