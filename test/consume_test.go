package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/tym_hmm/rabbitmq-pool-go"
	"sync"
	"testing"
)

func TestConsume(t *testing.T)  {
	initConsumerabbitmq()
	Consume()
}

func Consume() {
	nomrl := &kelleyRabbimqPool.ConsumeReceive{
		ExchangeName: "testChange31",//队列名称
		ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT,
		Route:        "",
		QueueName:    "testQueue31",
		IsTry:true,//是否重试
		IsAutoAck:false,//自动消息确认
		MaxReTry: 5,//最大重试次数
		EventFail: func(code int, e error, data []byte) {
			fmt.Printf("error:%s", e)
		},
		EventSuccess: func(data []byte, header map[string]interface{}, retryClient kelleyRabbimqPool.RetryClientInterface)bool {//如果返回true 则无需重试
			_ = retryClient.Ack()
			fmt.Printf("data:%s\n", string(data))
			return true
		},
	}
	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
}

var onceConsumePool sync.Once
var instanceConsumePool *kelleyRabbimqPool.RabbitPool

func initConsumerabbitmq() *kelleyRabbimqPool.RabbitPool {
	onceConsumePool.Do(func() {
		instanceConsumePool = kelleyRabbimqPool.NewConsumePool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		err := instanceConsumePool.Connect("192.168.1.169", 5672, "admin", "admin")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}