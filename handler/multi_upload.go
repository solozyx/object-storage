package handler

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	// redis驱动
	_ "github.com/garyburd/redigo/redis"
	"github.com/gomodule/redigo/redis"

	rPool "github.com/solozyx/object-storage/cache/redis"
	conf "github.com/solozyx/object-storage/config"
	"github.com/solozyx/object-storage/db"
	"github.com/solozyx/object-storage/util"
)

const (
	MPKeyPrefix      = conf.RdsCacheKeyPrefix + "MP_"
	ChunkCount       = "chunk_count"
	FileHash         = "file_hash"
	FileSize         = "file_size"
	ChunkIndexPrefix = "chunk_index_"
)

// 大文件分块上传元信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string // 唯一标识某1次上传操作
	ChunkSize  int    // 分块大小,最后1个分块要独立计算
	ChunkCount int    // 分块数量
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	// 2. 获得redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 生成分块上传的初始化元信息
	upInfo := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		// TODO:NOTICE 自定义规则 [用户名+当前时间戳]
		UploadID: username + fmt.Sprintf("%x", time.Now().UnixNano()),
		// 每个分块大小 5MB
		ChunkSize: 5 * 1024 * 1024,
		// 向上取整 math.Ceil()
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 4. 将初始化信息写入到redis缓存
	// TODO : NOTICE 使用 redis hash 这里可以使用 HMSET 一次性写入
	rConn.Do("HSET", MPKeyPrefix+upInfo.UploadID, ChunkCount, upInfo.ChunkCount)
	rConn.Do("HSET", MPKeyPrefix+upInfo.UploadID, FileHash, upInfo.FileHash)
	rConn.Do("HSET", MPKeyPrefix+upInfo.UploadID, FileSize, upInfo.FileSize)

	// 5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	// 当前分块序号
	chunkIndex := r.Form.Get("index")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// TODO : 文件分块哈希校验,每个分块要求客户端上传本地计算好的1个哈希值
	//  服务端接收到文件内容计算哈希,校验,文件分块完整继续

	// 3. 获得文件句柄 用于存储文件分块
	// 目录结构 /tmp/上传id/文件名是分块序号
	fStorePath := conf.FileLocalStorePath + uploadID + "/" + chunkIndex
	// TODO : NOTICE 创建文件,所在目录之前不存在,Create方法 no such file path error
	// fd, err := os.Create(fStorePath)
	// 权限 0744 当前用户拥有 7的权限 其他用户都是 4的只读权限
	os.MkdirAll(path.Dir(fStorePath), 0744)
	fd, err := os.Create(fStorePath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	// 存储用户上传的文件各个分块
	// 每次读取1MB文件数据 buffer
	buf := make([]byte, 1024*1024)
	for {
		// 读取文件数据到 buffer
		n, err := r.Body.Read(buf)
		// 将buffer文件数据写入存储句柄
		fd.Write(buf[:n])
		// 读取到Body最后 退出当前循环
		if err != nil {
			break
		}
	}

	// 4. 更新redis缓存状态 每上传完成1个文件分块增加1个记录 方便查询文件分块上传进度
	rConn.Do("HSET", MPKeyPrefix+uploadID, ChunkIndexPrefix+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 上传文件各个分块合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	uplodaId := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过uploadid查询redis 并判断 是否所有分块上传完成
	datas, err := redis.Values(rConn.Do("HGETALL", MPKeyPrefix+uplodaId))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}

	// 判断文件各个分块是否上传完毕
	totalCount := 0
	chunkCount := 0
	// TODO : NOTICE 通过 HGETALL 查询得到的结果 key value 在同1个数组中
	//  在每次循环中要同时解出 kv
	for i := 0; i < len(datas); i += 2 {
		k := string(datas[i].([]byte))
		v := string(datas[i+1].([]byte))
		if k == ChunkCount {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, ChunkIndexPrefix) && v == "1" {
			// k 以 chunk_index_ 开头 并且该分块上传完成 缓存值为 1
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		// 文件分块上传不完整
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// 4. TODO : 合并分块

	// 5. 更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	// 合并分块没做 fileaddr 传空字符串
	db.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 6. 响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
