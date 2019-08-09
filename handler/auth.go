package handler

import (
	"net/http"
)

// TODO : NOTICE 拦截器,类似python修饰器/java注解,原理类似,在目标函数执行入口前
//  先执行一段逻辑代码,通过后再执行目标函数
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// 解析前端Form表单参数
			r.ParseForm()
			username := r.Form.Get("username")
			token := r.Form.Get("token")
			//验证登录token是否有效
			if len(username) < 3 || !IsTokenValid(token) {
				// w.WriteHeader(http.StatusForbidden)
				// token校验失败则跳转到登录页面
				http.Redirect(w, r, "./static/view/signin.html", http.StatusFound)
				return
			}
			// token验证通过
			h(w, r)
		})
}
