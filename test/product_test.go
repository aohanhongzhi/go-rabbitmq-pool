package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/tym_hmm/rabbitmq-pool-go"
	"sync"
	"testing"
)

func TestProduct(t *testing.T) {
	initrabbitmq()
	rund()
}

var oncePool sync.Once
var instanceRPool *kelleyRabbimqPool.RabbitPool

func initrabbitmq() *kelleyRabbimqPool.RabbitPool {
	oncePool.Do(func() {
		instanceRPool = kelleyRabbimqPool.NewProductPool()
		//err := instanceRPool.Connect("192.168.1.169", 5672, "admin", "admin")
		err := instanceRPool.Connect("rabbitmq.cupb.top", 5672, "admin", "Jian,Yin.2019")

		//err:=instanceRPool.ConnectVirtualHost("192.168.1.169", 5672, "temptest", "test123456", "/temptest1")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}

func rund() {

	var wg sync.WaitGroup

	//wg.Add(1)
	//go func() {
	//	fmt.Println("aaaaaaaaaaaaaaaaaaaaaa")
	//	defer wg.Done()
	//	runtime.SetMutexProfileFraction(1)  // 开启对锁调用的跟踪
	//	runtime.SetBlockProfileRate(1)      // 开启对阻塞操作的跟踪
	//	err:= http.ListenAndServe("0.0.0.0:8080", nil)
	//	fmt.Println(err)
	//}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data := kelleyRabbimqPool.GetRabbitMqDataFormat("testChange31", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "testQueue31", "", fmt.Sprintf("这里是数据%d", num))
			err := instanceRPool.Push(data)
			if err != nil {
				fmt.Println(err)
			}
		}(i)
	}

	wg.Wait()
}
