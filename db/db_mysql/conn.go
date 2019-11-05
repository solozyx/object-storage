package db_mysql

import (
	// Go语言操作MySQL标准接口
	"database/sql"
	"fmt"
	"log"
	"os"

	// MySQL数据库驱动,一般不直接使用驱动提供的方法
	// 而是使用 "database/sql" 接口的 sql.DB
	// 所以匿名导入该驱动,导入 go-sql-driver/mysql 之后,该驱动进行初始化
	// 并且将驱动注册到 database/sql 的上下文
	// 就可以使用 database/sql 提供的接口方法操作MySQL数据库
	_ "github.com/go-sql-driver/mysql"

	conf "github.com/solozyx/object-storage/config"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", conf.MySQLSource)
	// 最大同时活跃连接数
	db.SetMaxOpenConns(1000)
	// 连接测试
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err:" + err.Error())
		os.Exit(1)
	}
}

// DBConn : 返回数据库连接对象
func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
