package config

const (
	// 是否开启文件异步转移(默认同步)
	AsyncTransferEnable = true

	// rabbitmq 服务入口url
	RabbitURL = "amqp://root:root@192.168.174.134:5672/object_storage"

	// 交换机 Exchange
	TransExchangeName = "exchange.object_storage.trans"
	// 队列 Queue
	TransOSSQueueName = "exchange.object_storage.trans.oss"
	// 路由键 RoutingKey
	TransOSSRoutingKey = "oss"

	//  oss转移失败后写入另一个队列
	TransOSSErrQueueName = "exchange.object_storage.trans.oss.err"
)
