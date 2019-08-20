package mq

import (
	"log"

	"github.com/streadway/amqp"
)

var done chan bool

// 接收消息
// callback 外部调用者指定的回调函数
func StartConsume(queueName, consumerName string, callback func(msg []byte) bool) {
	var(
		deliveries <-chan amqp.Delivery
		err error
	)
	// 消费者通过rabbitmq通信信道消费消息
	deliveries, err = c.Consume(
		// 队列名称
		queueName,
		// 消费者名称
		consumerName,
		// 自动应答,自动回复ACK,true 表示不用手动写代码回复已经收到消息的确认信号给生产者
		// rabbitmq客户端自动完成该项工作
		true,
		// 是否有多个消费者同时监听该Queue,rabbitmq server要根据一定竞争机制派发消息给不同的消费者
		// false表示 非唯一的消费者,可能有多个消费者监听该 Queue
		false,
		// rabbitmq不支持该参数,只能设置为false
		false,
		// noWait, false 表示客户端创建Queue的监听长连接后,会阻塞直到有消息过来
		false,
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	// 循环读取channel的数据
	done = make(chan bool)

	go func() {
		for delivery := range deliveries {
			processSucc := callback(delivery.Body)
			if !processSucc {
				// TODO : 将任务写入错误队列,待后续处理
			}
		}
	}()

	// 阻塞接收done信号 避免该函数退出
	<-done

	// 关闭rabbitmq通信信道
	c.Close()
}

// StopConsume : 停止监听队列
func StopConsume() {
	done <- true
}
