package main

import (
	"fmt"
	kelleyRabbimqPool "gitee.com/tym_hmm/rabbitmq-pool-go"
	nested "github.com/aohanhongzhi/nested-logrus-formatter"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"sync"
)

func main() {
	// 日志设置
	nested.LogInit()
	initrabbitmq()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		// 发送消息到mq
		value := c.Query("code")

		for i := 0; i < 10; i++ {
			if len(value) > 0 {
				Send(value)
			} else {
				Send("1")
			}
		}

		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // 监听并在 0.0.0.0:8080 上启动服务
}

var oncePool sync.Once
var instanceRPool *kelleyRabbimqPool.RabbitPool

func initrabbitmq() *kelleyRabbimqPool.RabbitPool {
	oncePool.Do(func() {
		instanceRPool = kelleyRabbimqPool.NewProductPool()
		//err := instanceRPool.Connect("192.168.1.169", 5672, "admin", "admin")
		err := instanceRPool.Connect("mysql.cupb.top", 5672, "admin", "Jian,Yin.2019")

		//err:=instanceRPool.ConnectVirtualHost("192.168.1.169", 5672, "temptest", "test123456", "/temptest1")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}

func Send(num string) {
	data := kelleyRabbimqPool.GetRabbitMqDataFormat("exHxb", kelleyRabbimqPool.EXCHANGE_TYPE_DIRECT, "", "route", fmt.Sprintf("这里是数据%v", num))
	err := instanceRPool.Push(data)
	if err != nil {
		log.Error(err)
	} else {
		log.Println("消息发送成功", num)
	}
}
