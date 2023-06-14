package aliyunOssService

import (
	"bytes"
	"com.copo/bo_service/boadmin/internal/svc"
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zeromicro/go-zero/core/logx"
)

func UploadFile(ctx context.Context, svctx *svc.ServiceContext, fileName string, bf *bytes.Buffer) (err error) {
	// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为https://oss-cn-hangzhou.aliyuncs.com。其它Region请按实际情况填写。
	endpoint := svctx.Config.Bucket.Host
	// 阿里云账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM用户进行API访问或日常运维，请登录RAM控制台创建RAM用户。
	accessKeyId := svctx.Config.Bucket.AccessKeyId
	accessKeySecret := svctx.Config.Bucket.AccessKeySecret
	// yourBucketName填写存储空间名称。
	bucketName := svctx.Config.Bucket.Name
	// yourObjectName填写Object完整路径，完整路径不包含Bucket名称。
	objectName := fileName

	// 创建OSSClient实例。
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logx.WithContext(ctx).Errorf("上传档案到阿里云错误，Err : '%v'", err.Error())

		return err
	}
	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logx.WithContext(ctx).Errorf("上传档案到阿里云错误，Err : '%v'", err.Error())

		return err
	}
	// 上传文件。
	err = bucket.PutObject(objectName, bytes.NewReader(bf.Bytes()))
	if err != nil {
		logx.WithContext(ctx).Errorf("上传档案到阿里云错误，Err : '%v'", err.Error())

		return err
	}

	return
}
