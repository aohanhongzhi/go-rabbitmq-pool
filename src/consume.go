package main

import (
	"fmt"
	"go-rabbitmq-pool/pool/rabbitmq"
	"sync"
)

func init() {

}
func main() {
	// 初始化
	initConsumerabbitmq()
	Consume()
}

func Consume() {
	nomrl := &rabbitmq.ConsumeReceive{

		ExchangeName: "testChange56",
		ExchangeType: rabbitmq.EXCHANGE_TYPE_TOPIC,
		Route:        "/",
		QueueName:    "testQueue56",
		EventFail: func(code int, e error) {
			//fmt.Printf("code:%d, error:%s",code,e)
		},
		EventSuccess: func(data []byte) {
			fmt.Printf("data:%s\n", string(data))
		},
	}
	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
}

var onceConsumePool sync.Once
var instanceConsumePool *rabbitmq.RabbitPool

func initConsumerabbitmq() *rabbitmq.RabbitPool {
	onceConsumePool.Do(func() {
		instanceConsumePool = rabbitmq.NewConsumePool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		err := instanceConsumePool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}
