package test

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/tym_hmm/rabbitmq-pool-go"
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

// 发送消息到队列
func TestSendQueueMessage(t *testing.T) {
	initrabbitmq()

	for i := 0; i < 100; i++ {
		//num := 1
		data := kelleyRabbimqPool.GetRabbitMqDataFormat("", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "jitu", "", fmt.Sprintf("这里是数据%d", i))
		err := instanceRPool.PushQueue(data)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(2 * time.Second)
	}

}
