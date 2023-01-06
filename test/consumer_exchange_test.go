package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/aohanhongzhi/go-rabbitmq-pool"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestConsumerExchange(t *testing.T) {
	initConsumerabbitmq()
	ConsumeExchange()
}

func ConsumeExchange() {
	nomrl := &kelleyRabbimqPool.ConsumeReceive{
		ExchangeName: "exHxb", // 交换机名称
		ExchangeType: kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT,
		Route:        "jt-wechat",
		QueueName:    "jt-wx-wechat",
		IsTry:        false, //是否重试
		IsAutoAck:    false, //自动消息确认
		MaxReTry:     5,     //最大重试次数
		EventFail: func(code int, e error, data []byte) {
			log.Printf("error:%s", e)
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
