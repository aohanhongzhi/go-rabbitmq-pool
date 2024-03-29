package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/aohanhongzhi/go-rabbitmq-pool"
	log "github.com/sirupsen/logrus"
	"sync"
	"testing"
)

func TestConsume(t *testing.T) {
	initConsumerabbitmq()
	Consume()
}

func Consume() {
	nomrl := &kelleyRabbimqPool.ConsumeReceive{
		ExchangeName: "testChange31", //交换机名称
		ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT,
		Route:        "route-exclusive",
		QueueName:    "test-return-exclusive",
		IsTry:        true,  //是否重试
		IsAutoAck:    false, //自动消息确认
		MaxReTry:     5,     //最大重试次数
		EventFail: func(code int, e error, data []byte) {
			fmt.Printf("error:%s", e)
		},
		EventSuccess: func(data []byte, header map[string]interface{}, retryClient kelleyRabbimqPool.RetryClientInterface) bool { //如果返回true 则无需重试
			_ = retryClient.Ack()
			log.Printf("data:%s", string(data))
			return true
		},
	}
	instanceConsumePool.RegisterConsumeReceive(nomrl)
	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	} else {
		log.Info("启动成功")
	}
}

var onceConsumePool sync.Once
var instanceConsumePool *kelleyRabbimqPool.RabbitPool

func initConsumerabbitmq() *kelleyRabbimqPool.RabbitPool {
	onceConsumePool.Do(func() {
		instanceConsumePool = kelleyRabbimqPool.NewConsumePool()
		//instanceConsumePool.SetMaxConsumeChannel(100)
		//err := instanceConsumePool.Connect("192.168.1.169", 5672, "admin", "admin")
		err := instanceConsumePool.Connect("mysql.cupb.top", 5672, "admin", "Jian,Yin.2019")
		//err:=instanceConsumePool.ConnectVirtualHost("192.168.1.169", 5672, "temptest", "test123456", "/temptest1")
		if err != nil {
			fmt.Println(err)
		} else {
			log.Info("监听消息连接池启动")
		}
		instanceConsumePool.SetMaxConsumeChannel(5)

	})
	return instanceConsumePool
}
