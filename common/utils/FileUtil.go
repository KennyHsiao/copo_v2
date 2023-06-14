package utils

import (
	"com.copo/bo_service/common/errorz"
	"com.copo/bo_service/common/random"
	"com.copo/bo_service/common/response"
	"github.com/copo888/transaction_service/common/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"mime/multipart"
	"os"
	"path"
	"strings"
)

// FileUpload acceptFileExts: 限制副檔名 ex : [".png",".jpg"]; folderPath: "./public/uploads/paytypeimgs/"
func FileUpload(media *multipart.FileHeader, acceptFileExts []string, folderPath string) (filePath string, err error) {

	ext := strings.ToLower(path.Ext(media.Filename))
	if utils.Contain(ext, acceptFileExts) {
		return "", errorz.New(response.FILE_EXTENSION_ERROR)
	}

	file, err := media.Open()
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		logx.Errorf("filed to read file from multipart: %s", err)
		return "", errorz.New(response.FILE_UPLOAD_ERROR)
	}

	newFileName := random.GetRandomString(12, random.ALL, random.MIX) + ext
	os.MkdirAll(folderPath, os.ModePerm)
	f, fileErr := os.OpenFile(folderPath+newFileName, os.O_WRONLY|os.O_CREATE, 0777)

	if fileErr != nil {
		logx.Error(fileErr.Error())
		return "", errorz.New(response.FILE_UPLOAD_ERROR, fileErr.Error())
	}
	defer f.Close()
	io.Copy(f, file)

	fullPath := f.Name()[1:len(f.Name())]
	filePath = strings.Replace(fullPath, "/public/", "", -1)

	return filePath, err
}
