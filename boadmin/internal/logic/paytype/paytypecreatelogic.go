package paytype

import (
	"com.copo/bo_service/boadmin/internal/config"
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
	"sort"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayTypeCreateLogic struct {
	logx.Logger
	Config config.Config
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayTypeCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) PayTypeCreateLogic {
	return PayTypeCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayTypeCreateLogic) PayTypeCreate(req types.PayTypeCreate) error {
	var imgUrl string
	//支付類型圖片上傳
	if req.UploadFile != nil {
		imgUrl2, err := l.PayTypeImageUpload(req)
		if err != nil {
			logx.Error("支付类型图片上传错误: ", err.Error())
		} else {
			//http://dev.res.copo.vip/bo_uploads/uploads/paytypeimgs/OrZfORpCVN.jpg
			//                           /public/uploads/paytypeimgs/OrZfORpCVN.jpg
			//imgUrl = l.svcCtx.Config.ResourceHost + "uploads/paytypeimgs/" + imgUrl2
			imgUrl = strings.Replace(imgUrl2, "/public/", "", -1)
			logx.Info("圖片位置:", imgUrl)
		}
	}

	return l.svcCtx.MyDB.Transaction(func(db *gorm.DB) (err error) {
		payType := &types.PayTypeCreate{
			PayTypeCreateRequest: req.PayTypeCreateRequest,
		}

		//判斷是否重複代碼及名字
		isDuplicated := model.NewPayType(db).CheckPayTypeDuplicated(req.PayTypeCreateRequest)
		if isDuplicated {
			logx.Error("支付類型重複: ", err)
			return errorz.New(response.PAY_TYPE_DUPLICATED)
		}
		if len(req.Currency) > 0 {
			var currencyStr = strings.Split(req.Currency, ",")
			sort.Strings(currencyStr)
			payType.Currency = strings.Join(currencyStr, ",")
		}

		if imgUrl != "" {
			payType.ImgUrl = imgUrl
			payType.PayTypeCreateRequest.ImgUrl = imgUrl
		}

		if err = l.svcCtx.MyDB.Table("ch_pay_types").Create(payType).Error; err != nil {
			return errorz.New(response.CREATE_FAILURE, err.Error())
		}

		return nil
	})
}

func (l *PayTypeCreateLogic) PayTypeImageUpload(req types.PayTypeCreate) (resp string, err error) {
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
