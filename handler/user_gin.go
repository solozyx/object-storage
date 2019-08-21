package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"

	cmn "filestore-server/common"
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/util"
)

// Gin分开处理http的get和post 处理注册GET请求响应注册页面
func Gin_SignupHandler(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "./static/view/signup.html")
}

// Gin 处理注册 POST 请求
func Gin_DoSignupHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	passwd := ctx.Request.FormValue("password")
	if len(username) < 3 || len(passwd) < 5 {
		// 200 表示服务端接收到客户端请求 和业务处理无关
		// TODO : NOTICE 通过Gin返回数据给客户端,除了返回静态页面内容的,其他返回都是JSON
		ctx.JSON(http.StatusOK ,gin.H{
			"msg":"Invalid parameter",
			"code":cmn.StatusParamInvalid,
		})
		return
	}

	// 对密码进行加盐 取Sha1值加密
	encPasswd := util.Sha1([]byte(passwd + cfg.UserSignupSalt))
	// 将用户注册信息写入用户表
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		ctx.JSON(http.StatusOK,gin.H{
			"msg":"Signup Success",
			"code":0,
		})
	} else {
		ctx.JSON(http.StatusOK,gin.H{
			"msg":"Signup Fail",
			"code":cmn.StatusRegisterFailed,
		})
	}
}

// Gin 处理登录 GET 请求 响应登录页面
func Gin_SignInHandler(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound,"./static/view/signin.html")
}

// Gin 处理登录 POST 请求
func Gin_DoSignInHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	password := ctx.Request.FormValue("password")
	encPasswd := util.Sha1([]byte(password + cfg.UserSignupSalt))

	// 1. 校验用户名及密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		ctx.JSON(http.StatusOK,gin.H{
			"msg":"Signin Fail",
			"code":cmn.StatusLoginFailed,
		})
		return
	}

	// 2. 生成访问凭证(token)下发给客户端
	token := GenToken(username)
	// token写入MySQL数据库
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		ctx.JSON(http.StatusOK,gin.H{
			"msg":"Signin Fail",
			"code":cmn.StatusTokenInvalid,
		})
		return
	}

	// 3. 登录成功后重定向到首页
	// w.Write([]byte("http://" + r.Host + "./static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		// 临时struct
		Data: struct {
			// 重定向url
			Location string
			Username string
			// 访问凭证
			Token    string
		}{
			Location: "./static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	ctx.Data(http.StatusOK,"application/json",resp.JSONBytes())
}

func Gin_UserInfoHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		ctx.JSON(http.StatusForbidden,gin.H{"msg":"Get Userinfo Fail","code":cmn.StatusUserNotExists})
		return
	}
	resp := util.RespMsg{Code: 0,Msg:  "OK",Data: user}
	ctx.Data(http.StatusOK,"application/json",resp.JSONBytes())
}
