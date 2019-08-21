package main

import (
	"filestore-server/handler"

	// "filestore-server/assets"
	conf "filestore-server/config"
	"fmt"
	"net/http"
)

func main() {
	// 静态资源处理
	// http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assets.AssetFS())))
	http.Handle("/static/",	http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 用户相关接口
	http.HandleFunc("/", handler.SignInHandler)
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)

	// 以下是需要中间件校验token的接口
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))
	// 文件存取接口
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/suc", handler.HTTPInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))
	// 秒传接口
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(handler.TryFastUploadHandler))
	// 上传文件下载接口
	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(handler.DownloadURLHandler))
	// 分块上传接口
	http.HandleFunc("/file/mpupload/init",handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart",handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete",handler.HTTPInterceptor(handler.CompleteUploadHandler))

	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", conf.UploadServiceHost)
	// 启动服务并监听端口
	err := http.ListenAndServe(conf.UploadServiceHost, nil)
	if err != nil {
		fmt.Printf("Failed to start server, err:%s", err.Error())
	}
}
