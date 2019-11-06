package handler

import (
	"fmt"
	"net/http"
	"time"

	conf "github.com/solozyx/object-storage/config"
	"github.com/solozyx/object-storage/db"
	"github.com/solozyx/object-storage/util"
)

// 处理用户注册请求
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	// GET
	if r.Method == http.MethodGet {
		// 用户注册页面
		// data, err := ioutil.ReadFile("./static/view/signup.html")
		// if err != nil {
		// 	 w.WriteHeader(http.StatusInternalServerError)
		// 	 return
		// }
		// w.Write(data)
		// return

		http.Redirect(w, r, "./static/view/signup.html", http.StatusFound)
		return
	}

	// POST
	r.ParseForm()
	username := r.Form.Get("username")
	passWd := r.Form.Get("password")
	if len(username) < 3 || len(passWd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}
	// 对密码进行加盐 取Sha1值加密
	encryptPassWd := util.Sha1([]byte(passWd + conf.UserSignupSalt))
	// 将用户注册信息写入用户表
	ok := db.UserSignUp(username, encryptPassWd)
	if ok {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// data, err := ioutil.ReadFile("./static/view/signin.html")
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// w.Write(data)
		http.Redirect(w, r, "./static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encryptPassWd := util.Sha1([]byte(password + conf.UserSignupSalt))

	// 1. 校验用户名及密码
	pwdChecked := db.UserSignIn(username, encryptPassWd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	// 2. 生成访问凭证(token)下发给客户端
	token := genToken(username)
	// token写入MySQL数据库
	upRes := db.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
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
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// UserInfoHandler : 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求表单参数
	r.ParseForm()
	username := r.Form.Get("username")

	// 验证token放到http请求拦截器
	// 2. 验证token是否有效
	token := r.Form.Get("token")
	ok := isTokenValid(token)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 3. 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// 生成token,规则自定义,这里生成 40字符长度的token
func genToken(username string) string {
	// MD5 字符串长度是 32位 + 当前时间戳的前8位
	// 40位字符 : md5(username + timestamp + token_salt) + timestamp[:8]
	// 当前时间戳
	ts := fmt.Sprintf("%x", time.Now().Unix())
	// MD5字符串
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	// 40位 token字符串
	return tokenPrefix + ts[:8]
}

// token是否有效
func isTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}
