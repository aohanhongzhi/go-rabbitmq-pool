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
<<<<<<< HEAD
		ExchangeName: "fn_fnout_test",
		ExchangeType: rabbitmq.EXCHANGE_TYPE_FANOUT,
		Route:        "fn_fnout_test_1",
		QueueName:    "fn_fnout_test_1",
=======
		ExchangeName: "testChange56",
		ExchangeType: rabbitmq.EXCHANGE_TYPE_TOPIC,
		Route:        "/",
		QueueName:    "testQueue56",
>>>>>>> 68159f7c0b8e480c6af1a1ece04cfca690bd1e75
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
<<<<<<< HEAD
		err := instanceConsumePool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
=======
		err := instanceConsumePool.Connect("47.108.223.220", 32002, "rabbituser", "rabbitpass")
>>>>>>> 68159f7c0b8e480c6af1a1ece04cfca690bd1e75
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsumePool
}
