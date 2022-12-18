package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/tym_hmm/rabbitmq-pool-go"
	"testing"
	"time"
)

func TestSendExchange(t *testing.T) {
	initrabbitmq()
	SendExchangeHxb(1)
	time.Sleep(3 * time.Second)
}

func SendExchangeHxb(num int) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("exHxb", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "jt-wechat", fmt.Sprintf("这里是数据%d", num))
	err := instanceRPool.Push(data)
	if err != nil {
		fmt.Println(err)
	}
}
