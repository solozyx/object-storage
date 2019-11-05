package oss

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"

	conf "github.com/solozyx/object-storage/config"
)

var ossCli *oss.Client

// 创建oss client对象
func Client() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	ossCli, err := oss.New(conf.OSSEndpoint, conf.OSSAccesskeyID, conf.OSSAccessKeySecret)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	return ossCli
}

// Bucket : 获取bucket存储空间
func Bucket() *oss.Bucket {
	cli := Client()
	if cli != nil {
		bucket, err := cli.Bucket(conf.OSSBucket)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		return bucket
	}
	return nil
}

// OSS文件下载:
// 1.SDK
// 2.API 通过OSS签名临时授权bucket里面的某个Object,生成签名下载url
func DownloadURL(objName string) string {
	// 授权下载url有效时间 1小时
	signedURL, err := Bucket().SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return signedURL
}

// 针对指定 bucket 设置生命周期规则
// bucket里面的object遵循bucket规则
func BuildLifecycleRule(bucketName string) {
	// 表示前缀为 test/ 的对象(Object 文件)距最后修改时间30天后过期
	ruleTest1 := oss.BuildLifecycleRuleByDays("rule1", "test/", true, 30)
	rules := []oss.LifecycleRule{ruleTest1}
	Client().SetBucketLifecycle(bucketName, rules)
}
