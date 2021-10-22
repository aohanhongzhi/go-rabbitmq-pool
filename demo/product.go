package demo

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
		instanceRPool = rabbitmq.NewProductPool()
		err := instanceRPool.Connect("192.168.1.80", 5672, "fnadmin", "Fn123456")
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

	for i:=0;i<100000; i++ {
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
