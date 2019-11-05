package mq

import (
	"log"

	"github.com/streadway/amqp"

	conf "github.com/solozyx/object-storage/config"
)

var (
	// rabbitmq 连接对象
	conn *amqp.Connection
	// rabbitmq 通信信道
	c *amqp.Channel
	// 如果异常关闭 会接收通知
	notifyClose chan *amqp.Error
)

func init() {
	// 是否开启上传文件异步转移OSS功能 开启时才初始化rabbitmq连接
	if !conf.AsyncTransferEnable {
		return
	}

	if initChannel() {
		c.NotifyClose(notifyClose)
	}

	// 断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				c = nil
				log.Printf("mq.init rabbitmq onNotifyChannelClosed: %+v\n", msg)
				initChannel()
			}
		}
	}()
}

// 创建rabbitmq通信信道
func initChannel() bool {
	// 防止通信信道重复创建
	if c != nil {
		return true
	}

	// 创建rabbitmq连接
	conn, err := amqp.Dial(conf.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 创建rabbitmq通信信道
	c, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}

// 生产者 发布消息 生产者把消息投递到rabbitmq的交换机exchange
func Publish(exchange, routingKey string, msg []byte) bool {
	// 创建rabbitmq通信信道
	if !initChannel() {
		return false
	}
	// 发布消息到交换机
	err := c.Publish(
		exchange,   // 交换机
		routingKey, // 路由键
		false,      // 告诉交换机,如果该消息没有可以转发的Queue,就丢弃该消息
		false,      // 新版本rabbitmq该参数废弃,不起作用
		amqp.Publishing{ContentType: "text/plain", Body: msg})

	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}
