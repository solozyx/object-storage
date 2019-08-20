package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cmn "filestore-server/common"
	conf "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
)

const (
	CephKeyPrefix = "/ceph/"
	CephBucket = "userfile"

	// 注意:OSS上传文件不能以 / 开头 如 不能用 /oss/
	OSSKeyPrefix = "oss/"
)


// UploadHandler : 处理文件上传
// param w 向用户返回数据的RW对象
// param r 接收用户请求Request对象指针
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
		// 另一种返回方式:
		// 动态文件使用http.HandleFunc设置，静态文件使用到http.FileServer设置
		// 所以直接redirect到http.FileServer所配置的url
		// http.Redirect(w, r, "/static/view/index.html",  http.StatusFound)
	} else if r.Method == "POST" {
		// 接收文件流 存储到本地目录
		// 客户端浏览器Form表单提交文件
		// 文件句柄 文件头 错误信息
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data, err:%s\n", err.Error())
			return
		}
		// 函数退出前关闭文件句柄
		defer file.Close()
		
		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			// TODO : ERROR 不同用户同时上传相同文件名的文件,导致冲突
			Location: conf.FileLocalStorePath + head.Filename,
			UploadAt: time.Now().Format(conf.SysTimeform),
		}

		// 创建本地文件 接收文件流
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("Failed to create file, err:%s\n", err.Error())
			return
		}
		// 函数退出前关闭文件句柄
		defer newFile.Close()

		// 将内存文件流 拷贝到 本地文件buffer区 (目标文件,内存文件流) 返回 写入字节长度
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err:%s\n", err.Error())
			return
		}
		// 新文件句柄seek移动到文件最前面
		newFile.Seek(0, 0)
		// 计算文件哈希 Sha1 值
		// 上传超大文件计算文件哈希时间比较长,影响上传速度
		// 可以抽取为1个独立微服务,进行异步处理
		fileMeta.FileSha1 = util.FileSha1(newFile)

		// 游标重新回到文件头部
		newFile.Seek(0, 0)

		if conf.CurrentStoreType == cmn.StoreCeph {
			// 文件写入Ceph存储
			data, _ := ioutil.ReadAll(newFile)
			// 文件sha1保证用户上传文件存储在ceph的key的唯一性
			cephPath := CephKeyPrefix + fileMeta.FileSha1
			_ = ceph.PutObject(CephBucket, cephPath, data)
			// 更新文件路径,从本地存储路径 迁移为 Ceph云存储路径
			fileMeta.Location = cephPath
		} else if conf.CurrentStoreType == cmn.StoreOSS {
			// 文件写入OSS存储
			ossPath := OSSKeyPrefix + fileMeta.FileSha1
			// 判断写入OSS为同步还是异步
			if !conf.AsyncTransferEnable {
				err = oss.Bucket().PutObject(ossPath, newFile)
				if err != nil {
					fmt.Println(err.Error())
					w.Write([]byte("Upload OSS failed!"))
					return
				}
				fileMeta.Location = ossPath
			} else {
				// 写入异步转移任务队列
				data := mq.TransferData{
					FileHash:      fileMeta.FileSha1,
					CurLocation:   fileMeta.Location,
					DestLocation:  ossPath,
					DestStoreType: cmn.StoreOSS,
				}
				pubData, _ := json.Marshal(data)
				pubSuc := mq.Publish(
					conf.TransExchangeName,
					conf.TransOSSRoutingKey,
					pubData,
				)
				if !pubSuc {
					// TODO: 当前发送转移信息失败，稍后重试
				}
			}
		}

		// meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "./static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// UploadSucHandler : 上传已完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler : 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	// 解析浏览器表单参数
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	// fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if fMeta != nil {
		data, err := json.Marshal(fMeta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else {
		w.Write([]byte(`{"code":-1,"msg":"no such file"}`))
	}
}

// FileQueryHandler : 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	// fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 解析用户上传Form表单
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm, _ := meta.GetFileMetaDB(fsha1)

	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	// 小文件可以通过 ioutil 直接读取,大文件要用流方式
	// 直接把文件读取到内存中
	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 返回客户端 设置http响应头 让浏览器能识别
	w.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	w.Header().Set("content-disposition", "attachment; filename=\"" 	+ fm.FileName + "\"")
	w.Write(data)
}

// FileMetaUpdateHandler : 更新文件元信息接口(这里只实现重命名文件名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	if opType != "0" {
		// 403
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		// 405
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	curFileMeta := meta.GetFileMeta(fileSha1)
	// 更新文件名
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	// TODO : 更新文件表中的元信息记录

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// TODO : NOTICE 物理删除可能失败,认为删除了meta信息就是删除了文件
// FileDeleteHandler : 删除文件元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	fMeta := meta.GetFileMeta(fileSha1)
	// 磁盘物理上删除文件
	os.Remove(fMeta.Location)
	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)
	// TODO: 删除表文件信息
	w.WriteHeader(http.StatusOK)
}

// TryFastUploadHandler : 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 上传过则将文件信息写入用户文件表， 返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}

	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

// 文件下载接口,返回文件下载url而不是返回文件本身
// 对于上传到OSS的文件减少本地服务压力,减少1次网络传输
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	// 解析请求参数
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)

	// TODO: 判断文件存在OSS 还是Ceph 还是在本地
	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		username := r.Form.Get("username")
		token := r.Form.Get("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			r.Host, filehash, username, token)
		w.Write([]byte(tmpUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		// TODO: ceph下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		w.Write([]byte(signedURL))
	}
}
