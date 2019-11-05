package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	conf "github.com/solozyx/object-storage/config"
	"github.com/solozyx/object-storage/handler"
)

func main() {
	// Gin Framework 创建包括 Logger Recovery 中间件的路由器
	// 使用 gin.New()创建的路由不包括中间件
	router := gin.Default()

	// 处理静态资源
	router.Static("/static/","./static")

	// 无需鉴权接口
	// 登录
	router.GET("/", handler.Gin_DoSignInHandler)
	router.GET("/user/sign_in", handler.Gin_SignInHandler)
	router.POST("/user/sign_in", handler.Gin_DoSignInHandler)
	// 注册
	router.GET("/user/sign_up", handler.Gin_SignUpHandler)
	router.POST("/user/sign_up", handler.Gin_DoSignUpHandler)

	// 中间件,校验token
	router.Use(handler.Gin_HTTPInterceptor())

	// TODO : NOTICE 在Use方法之后的所有 handler 都会经过拦截器进行token校验
	router.GET("/user/info",handler.Gin_UserInfoHandler)

	// ...

	// 启动服务并监听端口
	fmt.Printf("上传服务启动中，开始监听监听[%s]...\n", conf.UploadServiceHost)
	router.Run(conf.UploadServiceHost)
}
