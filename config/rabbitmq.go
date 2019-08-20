package config

const (
	// 是否开启文件异步转移(默认同步)
	AsyncTransferEnable = true

	// rabbitmq 服务入口url
	RabbitURL = "amqp://guest:guest@127.0.0.1:5672/"

	// 交换机 Exchange
	TransExchangeName = "uploadserver.trans"
	// 队列 Queue
	TransOSSQueueName = "uploadserver.trans.oss"
	// 路由键 RoutingKey
	TransOSSRoutingKey = "oss"

	//  oss转移失败后写入另一个队列
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
)
