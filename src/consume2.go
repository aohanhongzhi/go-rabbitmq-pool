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
	initConsume2rabbitmq()
	Consume2()
}

func Consume2() {
	nomrl := &rabbitmq.ConsumeReceive{
		ExchangeName: "fn_fnout_test",
		ExchangeType: rabbitmq.EXCHANGE_TYPE_FANOUT,
		Route:        "fn_fnout_test_2",
		QueueName:    "fn_fnout_test_2",
		EventFail: func(code int, e error) {
			//fmt.Printf("code:%d, error:%s",code,e)
		},
		EventSuccess: func(data []byte) {
			fmt.Printf("data:%s\n", string(data))
		},
	}
	instanceConsume2Pool.RegisterConsumeReceive(nomrl)
	err := instanceConsume2Pool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
}

var onceConsume2Pool sync.Once
var instanceConsume2Pool *rabbitmq.RabbitPool

func initConsume2rabbitmq() *rabbitmq.RabbitPool {
	onceConsume2Pool.Do(func() {
		instanceConsume2Pool = rabbitmq.NewRabbitPool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		err := instanceConsume2Pool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceConsume2Pool
}
