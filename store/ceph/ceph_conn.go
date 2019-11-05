package ceph

import (
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"

	conf "github.com/solozyx/object-storage/config"
)

var cephConn *s3.S3

// 获取ceph连接
func GetCephConnection() *s3.S3 {
	// 防止重复初始化连接
	if cephConn != nil {
		return cephConn
	}
	// 1. 初始化ceph
	auth := aws.Auth{
		AccessKey: conf.CephAccessKey,
		SecretKey: conf.CephSecretKey,
	}
	// 2. 创建S3类型连接
	// 设置Region地区
	curRegion := aws.Region{
		Name: "default",
		// ceph集群 radosgw 网关服务地址
		EC2Endpoint: conf.CephGWEndpoint,
		// ceph集群 radosgw 网关服务地址
		S3Endpoint: conf.CephGWEndpoint,
		// 不用指定
		S3BucketEndpoint: "",
		// 这里不需要做区域限制
		S3LocationConstraint: false,
		// false表示 创建的bucket允许大小写
		S3LowercaseBucket: false,
		// 签名算法
		Sign: aws.SignV2,
	}
	return s3.New(auth, curRegion)
}

// 获取指定的bucket对象
func GetCephBucket(bucket string) *s3.Bucket {
	conn := GetCephConnection()
	return conn.Bucket(bucket)
}

// 上传文件到ceph集群
func PutObject(bucket string, path string, data []byte) error {
	// TODO : NOTICE 根据业务选择权限级别 这里默认使用 PublicRead
	return GetCephBucket(bucket).Put(path, data, "octet-stream", s3.PublicRead)
}
