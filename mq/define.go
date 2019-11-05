package mq

import (
	comm "github.com/solozyx/object-storage/common"
)

// 将要写到rabbitmq的数据的结构体
type TransferData struct {
	FileHash      string
	CurLocation   string         // 上传文件本地存储路径
	DestLocation  string         // 文件迁移到OSS的目标地址
	DestStoreType comm.StoreType // 文件迁移到存储服务类型
}
