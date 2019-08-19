package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	pool      *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPass = "" // 开启Redis验证可以设置登录密码 testupload
)

// 创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		// 连接池最多有多少个可用连接
		MaxIdle:     50,
		// 同时能够使用的连接数 0表示无限制 <= MaxIdle
		MaxActive:   30,
		// 1个连接超过 IdleTimeout 都没有使用则回收该连接
		IdleTimeout: 300 * time.Second,
		// 创建配置1个连接 返回1个连接对象
		Dial: func() (redis.Conn, error) {
			// 1. 打开连接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// 2. 访问认证
			if _, err = c.Do("AUTH", redisPass); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		// 定时检查连接是否可用 redis-server 健康状况
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			// 每分钟检测1次 redis-server 健康状态
			if time.Since(t) < time.Minute {
				// 小于1分钟不做检测
				return nil
			}
			// 超过1分钟 PING 命令检测
			_, err := conn.Do("PING")
			// 有错误 连接池会自动关闭
			return err
		},
	}
}

// 程序启动 初始化redis连接池
func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {
	// pool在package外访问不到,对外方法供外部访问
	return pool
}
