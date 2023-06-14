package paytype

import (
	"com.copo/bo_service/boadmin/internal/model"
	"com.copo/bo_service/boadmin/internal/svc"
	"com.copo/bo_service/boadmin/internal/types"
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"context"
	"gorm.io/gorm"
	"io"
	"os"
	"path"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeUpdateLogic {
	return PayTypeUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeUpdateLogic) PayTypeUpdate(req types.PayTypeUpdate) error {
	var imgUrl string
	//支付類型圖片上傳
	if req.UploadFile != nil {
		imgUrl2, err := l.PayTypeImageUpload(req)
		if err != nil {
			logx.Error("支付类型图片上传错误: ", err.Error())
		} else {
			//imgUrl = l.svcCtx.Config.ResourceHost + "uploads/paytypeimgs/" + imgUrl2
			imgUrl = strings.Replace(imgUrl2, "/public/", "", -1)
			logx.Info("圖片位置:", imgUrl)
		}
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		payType := &types.PayTypeUpdate{
			PayTypeUpdateRequest: req.PayTypeUpdateRequest,
		}
		//判斷是否存在
		if err = model.NewPayType(db).CheckPayTypeExist(req.PayTypeUpdateRequest); err != nil {
			return errorz.New(response.PAY_TYPE_NOT_EXIST)
		}
		if imgUrl != "" {
			payType.ImgUrl = imgUrl
			payType.PayTypeUpdateRequest.ImgUrl = imgUrl
		}

		if err = l.svcCtx.MyDB.Table("ch_pay_types").Updates(payType).Error; err != nil {
			return errorz.New(response.UPDATE_DATABASE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *PayTypeUpdateLogic) PayTypeImageUpload(req types.PayTypeUpdate) (resp string, err error) {
	var file = req.UploadFile
	var header = req.UploadHeader
	ext := strings.ToLower(path.Ext(header.Filename))
	if ext != ".jpg" && ext != ".png" {
		return "", errorz.New(response.FILE_TYPE_NOT_JPG_ERROR)
	}
	defer file.Close()
	var terms []string
	randStr := random.GetRandomString(10, random.ALL, random.MIX)
	terms = append(append(terms, randStr), ext)
	newFileName := strings.Join(terms, "")
	f, errOpenFile := os.OpenFile("./public/uploads/paytypeimgs/"+newFileName, os.O_WRONLY|os.O_CREATE, 0777)
	//f, errOpenFile := os.OpenFile("C:\\public\\uploads\\internalcharges\\"+newFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if errOpenFile != nil {
		logx.Error(errOpenFile.Error())
		return "", errorz.New(response.FAIL, errOpenFile.Error())
	}
	defer f.Close()
	//把.去掉
	//splitStr := strings.Split(f.Name(), ".")
	fileName := f.Name()[1:len(f.Name())]
	io.Copy(f, file)
	return fileName, nil
}
