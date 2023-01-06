package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/aohanhongzhi/go-rabbitmq-pool"
	log "github.com/sirupsen/logrus"
	"sync"
	"testing"
	"time"
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
		//err := instanceRPool.Connect("mysql.cupb.top", 5672, "admin", "Jian,Yin.2019")
		err := instanceRPool.Connect("mysql.cupb.top", 5672, "admin", "Jian,Yin.2019")

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
			Send(num)
		}(i)
	}
	wg.Wait()
}

func Send(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("testChange31", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "route", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		fmt.Println(err)
	}
}
func Sendexclusive(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("testChange31", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "route-exclusive", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		fmt.Println(err)
	}
}
func SendexHxb(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("exHxb", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "route-exclusive", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		fmt.Println(err)
	}
}

func SendexHxb1(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("exHxb", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "xiaobai_one", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		log.Error(err)
	}
}

func SendexHxbTwo(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("exHxb", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "xiaobai_two", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		log.Error(err)
	}
}

func TestSendOne(t *testing.T) {
	initrabbitmq()
	Send(1)
}
func TestSendexclusive(t *testing.T) {
	initrabbitmq()
	Sendexclusive(231)
	time.Sleep(10 * time.Second)
}
func TestSendexHxb(t *testing.T) {
	initrabbitmq()
	SendexHxb(231)
	time.Sleep(1 * time.Second)
}

func TestSendexHxbTwo(t *testing.T) {
	initrabbitmq()
	SendexHxb(231)
	SendexHxbTwo(2321)
	time.Sleep(3 * time.Second)
}
