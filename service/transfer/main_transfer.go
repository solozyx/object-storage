package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"

	conf "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
)

// 上传文件从本地存储 迁移 阿里云OSS存储
func ProcessTransfer(msg []byte) bool {
	data := mq.TransferData{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 本地临时文件存储路径,创建文件句柄
	fileLocal, err := os.Open(data.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 读取本地文件,上传OSS
	err = oss.Bucket().PutObject(data.DestLocation,bufio.NewReader(fileLocal))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 更新文件存储信息到文件表
	succ := dblayer.UpdateFileLocation(data.FileHash,data.DestLocation)
	if !succ {
		return false
	}
	return true
}

func main() {
	if !conf.AsyncTransferEnable {
		log.Println("上传文件从本地存储异步转移阿里云OSS存储,目前被禁用,请检查相关配置")
		return
	}
	log.Println("上传文件从本地存储异步转移阿里云OSS存储,开始监听rabbitmq转移任务队列 ... ")
	mq.StartConsume(conf.TransOSSQueueName,"transfer_oss",ProcessTransfer)
}
