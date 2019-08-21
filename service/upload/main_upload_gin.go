package main

import (
	"fmt"
	"github.com/gin-gonic/gin"

	// "filestore-server/assets"
	conf "filestore-server/config"
	"filestore-server/handler"
)

func main() {
	// Gin Framework 创建包括 Logger Recovery 中间件的路由器
	// 使用 gin.New()创建的路由不包括中间件
	router := gin.Default()

	// 处理静态资源
	router.Static("/static/","./static")

	// 无需鉴权接口
	// 登录
	router.GET("/", handler.Gin_SignupHandler)
	router.GET("/user/signin", handler.Gin_SignInHandler)
	router.POST("/user/signin", handler.Gin_DoSignInHandler)
	// 注册
	router.GET("/user/signup", handler.Gin_SignupHandler)
	router.POST("/user/signup", handler.Gin_DoSignupHandler)

	// 中间件,校验token
	router.Use(handler.Gin_HTTPInterceptor())

	// TODO : NOTICE 在Use方法之后的所有 handler 都会经过拦截器进行token校验
	router.GET("/user/info",handler.Gin_UserInfoHandler)
	// router.GET("/file/upload",handler.Gin_UploadHandler)
	// router.GET("/file/upload",handler.Gin_DoUploadHandler)
	// router.GET("/file/upload/suc",handler.Gin_UploadSucHandler)
	// router.GET("/file/meta",handler.Gin_GetFileMetaHandler)
	// router.GET("/file/query",handler.Gin_FileQueryHandler)
	// router.GET("/file/download",handler.Gin_DownloadHandler)
	// router.GET("/file/update",handler.Gin_FileMetaUpdateHandler)
	// router.GET("/file/delete",handler.Gin_FileDeleteHandler)
	// router.GET("/file/fastupload",handler.Gin_TryFastUploadHandler)
	// ...

	// 启动服务并监听端口
	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", conf.UploadServiceHost)
	router.Run(conf.UploadServiceHost)
}
