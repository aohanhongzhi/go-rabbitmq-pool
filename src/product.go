package main

import (
	"fmt"
	"go-rabbitmq-pool/pool/rabbitmq"
	_ "net/http/pprof"
	"sync"
)
func init() {
	// 初始化
	initrabbitmq()
}

func main() {
	fmt.Println("开始")
	rund()
}

var oncePool sync.Once
var instanceRPool *rabbitmq.RabbitPool

func initrabbitmq() *rabbitmq.RabbitPool {
	oncePool.Do(func() {
<<<<<<< HEAD
		instanceRPool = rabbitmq.NewRabbitPool()
		err := instanceRPool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
=======
		instanceRPool = rabbitmq.NewProductPool()
		//47.108.223.220:32002
		err := instanceRPool.Connect("47.108.223.220", 32002, "rabbituser", "rabbitpass")
>>>>>>> 68159f7c0b8e480c6af1a1ece04cfca690bd1e75
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}

func rund() {

	var wg sync.WaitGroup
<<<<<<< HEAD

	//wg.Add(1)
	//go func() {
	//	fmt.Println("aaaaaaaaaaaaaaaaaaaaaa")
	//	defer wg.Done()
	//	runtime.SetMutexProfileFraction(1)  // 开启对锁调用的跟踪
	//	runtime.SetBlockProfileRate(1)      // 开启对阻塞操作的跟踪
	//	err:= http.ListenAndServe("0.0.0.0:8080", nil)
	//	fmt.Println(err)
	//}()

	for i:=0;i<100000; i++ {
=======
	for i:=0;i<500000; i++ {
>>>>>>> 68159f7c0b8e480c6af1a1ece04cfca690bd1e75
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data:=rabbitmq.GetRabbitMqDataFormat("fn_fnout_test", rabbitmq.EXCHANGE_TYPE_FANOUT, "", "", fmt.Sprintf("这里是数据%d", num))
			err:=instanceRPool.Push(data)
			if err!=nil{
				fmt.Println(err)
			}
		}(i)
	}

	wg.Wait()
}
