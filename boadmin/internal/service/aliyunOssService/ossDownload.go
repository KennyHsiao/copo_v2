package aliyunOssService

import (
	"com.copo/bo_service/boadmin/internal/svc"
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"
)

func DownloadURL(ctx context.Context, svctx *svc.ServiceContext, fileName string) (resp string, err error) {
	// yourEndpoint填写Bucket对应的Endpoint，以华东1（杭州）为例，填写为https://oss-cn-hangzhou.aliyuncs.com。其它Region请按实际情况填写。
	endpoint := svctx.Config.Bucket.Host
	// 阿里云账号AccessKey拥有所有API的访问权限，风险很高。强烈建议您创建并使用RAM用户进行API访问或日常运维，请登录RAM控制台创建RAM用户。
	accessKeyId := svctx.Config.Bucket.AccessKeyId
	accessKeySecret := svctx.Config.Bucket.AccessKeySecret
	// yourBucketName填写存储空间名称。
	bucketName := svctx.Config.Bucket.Name
	// yourObjectName填写Object完整路径，完整路径不包含Bucket名称。
	objectName := fileName
	// 获取STS临时凭证后，您可以通过其中的安全令牌（SecurityToken）和临时访问密钥（AccessKeyId和AccessKeySecret）生成OSSClient。
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		logx.WithContext(ctx).Errorf("取得下载url失败，Err : '%v'", err.Error())
		return "", err
	}

	// 将Object下载到本地文件，并保存到指定的本地路径中。如果指定的本地文件存在会覆盖，不存在则新建。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logx.WithContext(ctx).Errorf("取得下载url失败，Err : '%v'", err.Error())
		return "", err
	}

	// 生成用于下载的签名URL，并指定签名URL的有效时间为30天。
	signedURL, errSignURL := bucket.SignURL(objectName, oss.HTTPGet, 2592000)
	if errSignURL != nil {
		logx.WithContext(ctx).Errorf("取得下载url失败，Err : '%v'", err.Error())
		return "", err
	}

	httpsSignedURL := strings.Replace(signedURL, "http", "https", 1)
	return httpsSignedURL, nil
}
