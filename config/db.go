package config

const (
	// MySQLSource 要连接的数据库源
	// 其中 root:root 用户名密码
	// fileserver 数据库名
	// charset=utf8 指定了数据以utf8字符编码进行传输
	MySQLSource = "root:root@tcp(192.168.174.134:3306)/object_storage?charset=utf8"
)
