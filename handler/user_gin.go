package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	comm "github.com/solozyx/object-storage/common"
	conf "github.com/solozyx/object-storage/config"
	"github.com/solozyx/object-storage/db"
	"github.com/solozyx/object-storage/util"
)

// Gin分开处理http的get和post 处理注册GET请求响应注册页面
func Gin_SignUpHandler(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "./static/view/signup.html")
}

// Gin 处理注册 POST 请求
func Gin_DoSignUpHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	passwd := ctx.Request.FormValue("password")
	if len(username) < 3 || len(passwd) < 5 {
		// 200 表示服务端接收到客户端请求 和业务处理无关
		// TODO : NOTICE 通过Gin返回数据给客户端,除了返回静态页面内容的,其他返回都是JSON
		ctx.JSON(http.StatusOK, gin.H{
			"msg":  "Invalid parameter",
			"code": comm.StatusParamInvalid,
		})
		return
	}

	// 对密码进行加盐 取Sha1值加密
	encryptPassWd := util.Sha1([]byte(passwd + conf.UserSignupSalt))
	// 将用户注册信息写入用户表
	ok := db.UserSignUp(username, encryptPassWd)
	if ok {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":  "SignUp Success",
			"code": 0,
		})
	} else {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":  "SignUp Fail",
			"code": comm.StatusRegisterFailed,
		})
	}
}

// Gin 处理登录 GET 请求 响应登录页面
func Gin_SignInHandler(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "./static/view/signin.html")
}

// Gin 处理登录 POST 请求
func Gin_DoSignInHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	password := ctx.Request.FormValue("password")
	encryptPassWd := util.Sha1([]byte(password + conf.UserSignupSalt))

	// 1. 校验用户名及密码
	pwdChecked := db.UserSignIn(username, encryptPassWd)
	if !pwdChecked {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":  "SignIn Fail",
			"code": comm.StatusLoginFailed,
		})
		return
	}

	// 2. 生成访问凭证(token)下发给客户端
	token := genToken(username)
	// token写入MySQL数据库
	ok := db.UpdateToken(username, token)
	if !ok {
		ctx.JSON(http.StatusOK, gin.H{
			"msg":  "SignIn Fail",
			"code": comm.StatusTokenInvalid,
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
			Token string
		}{
			Location: "./static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	ctx.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

func Gin_UserInfoHandler(ctx *gin.Context) {
	username := ctx.Request.FormValue("username")
	user, err := db.GetUserInfo(username)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"msg": "Get UserInfo Fail", "code": comm.StatusUserNotExists})
		return
	}
	resp := util.RespMsg{Code: 0, Msg: "OK", Data: user}
	ctx.Data(http.StatusOK, "application/json", resp.JSONBytes())
}
