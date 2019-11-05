package db

import (
	"fmt"

	"github.com/solozyx/object-storage/db/db_mysql"
)

// User : 数据库 tbl_user 用户表model模型
type User struct {
	Username     string
	Email        string
	Phone        string
	SignUpAt     string
	LastActiveAt string
	Status       int
}

// 通过用户名及密码完成user表的注册操作
func UserSignUp(username string, passWd string) bool {
	stmt, err := db_mysql.DBConn().Prepare("insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, passWd)
	if err != nil {
		fmt.Println("Failed to insert, err:" + err.Error())
		return false
	}
	// SQL执行成功,判断是否真正插入了数据,重复注册在逻辑上也算失败
	if rowsAffected, err := ret.RowsAffected(); nil == err && rowsAffected > 0 {
		return true
	}
	return false
}

// 用户登录,判断密码是否一致
func UserSignIn(username string, encpwd string) bool {
	stmt, err := db_mysql.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found: " + username)
		return false
	}

	pRows := db_mysql.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

//  更新用户登录的token
func UpdateToken(username string, token string) bool {
	stmt, err := db_mysql.DBConn().Prepare("replace into tbl_user_token (`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 查询用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := db_mysql.DBConn().Prepare("select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignUpAt)
	if err != nil {
		return user, err
	}
	return user, nil
}
