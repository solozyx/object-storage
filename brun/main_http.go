package main

import (
	"fmt"
	"net/http"

	conf "github.com/solozyx/object-storage/config"
	"github.com/solozyx/object-storage/handler"
)

func main() {
	// 静态资源处理
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 用户相关接口
	http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/sign_up", handler.SignUpHandler)
	http.HandleFunc("/user/sign_in", handler.SignInHandler)

	// 以下是需要中间件校验token的接口
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	// 文件普通上传 存取接口
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/success", handler.HTTPInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	// 上传文件 本地存储 下载接口
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))

	// 文件秒传接口
	http.HandleFunc("/file/fast_upload", handler.HTTPInterceptor(handler.TryFastUploadHandler))
	// 上传文件 云存储 下载接口
	http.HandleFunc("/file/download_url", handler.HTTPInterceptor(handler.DownloadURLHandler))
	// 分块上传接口
	http.HandleFunc("/file/mp_upload/init", handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mp_upload/upload_part", handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mp_upload/complete", handler.HTTPInterceptor(handler.CompleteUploadHandler))
	// http.HandleFunc("/file/mp_upload/cancel"
	// http.HandleFunc("/file/mp_upload/status"

	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", conf.UploadServiceHost)
	// 启动服务并监听端口
	err := http.ListenAndServe(conf.UploadServiceHost, nil)
	if err != nil {
		fmt.Printf("Failed to start server, err:%s", err.Error())
	}
}
