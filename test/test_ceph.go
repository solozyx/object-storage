package main

import (
	"fmt"
	"os"

	"gopkg.in/amz.v1/s3"

	"filestore-server/store/ceph"
)

func main() {
	ceph2Local()
	return

	bucket := ceph.GetCephBucket("userfile")

	// 创建一个新的bucket
	// PublicRead 所有用户都可以访问上传的对象,无需额外验证
	err := bucket.PutBucket(s3.PublicRead)
	fmt.Printf("create bucket err: %v\n", err)

	// 查询这个bucket下面指定条件的 object keys
	res, _ := bucket.List("", "", "", 100)
	fmt.Printf("object keys: %+v\n", res)

	// 新上传一个对象
	// path 类似文件的绝对路径
	// 上传文件的具体内容
	// 文件类型
	// 权限
	err = bucket.Put("/testupload/a.txt",
		[]byte("just for test"),
		"octet-stream",
		s3.PublicRead)
	fmt.Printf("upload err: %+v\n", err)

	// 查询这个bucket下面指定条件的object keys
	res, err = bucket.List("", "", "", 100)
	fmt.Printf("object keys: %+v\n", res)
}

func ceph2Local(){
	bucket := ceph.GetCephBucket("userfile")
	// 从ceph读取1个用户上传的文件
	d, _ := bucket.Get("/ceph/866cc7c87c9b612dd8904d2c5dd07d6f6c22b834")
	// 把文件恢复到本地磁盘
	tmpFile, _ := os.Create("/tmp/test_file")
	tmpFile.Write(d)
}