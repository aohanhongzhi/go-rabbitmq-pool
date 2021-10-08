## rabbitmq 连接池channel复用

开发语言 golang
依赖库

require github.com/streadway/amqp v1.0.0
```
go get -u github.com/streadway/amqp v1.0.0
或
go mod tidy
```

> 已在线上生产环镜运行， 5200W请求 qbs 3000 时， 连接池显示无压力<br>
> rabbitmq部署为线上集群

### 功能说明
1. 自定义连接池大小及最大处理channel数
2. 消费者底层断线自动重连
3. 底层使用轮循方式复用tcp
4. 生产者每个tcp对应一个channel,防止channel写入阻塞造成内存使用过量
5. 支持rabbitmq exchangeType
6. 默认值

| 名称 | 说明 |
| - | - |
| tcp最大连接数 | 5 |
| 生产者消费发送失败最大重试次数 | 5 |
| 消费者最大channel信道数(每个连接自动平分) | 100(每个tcp10个) |



### 使用
1. 初始化
```
var oncePool sync.Once
var instanceRPool *rabbitmq.RabbitPool
func initrabbitmq() *rabbitmq.RabbitPool {
	oncePool.Do(func() {
        //初始化生产者
		instanceRPool = rabbitmq.NewProductPool()
        //初始化消费者
	    instanceConsumePool = rabbitmq.NewConsumePool()
		err := instanceRPool.Connect("192.168.1.202", 5672, "guest", "guest")
		if err != nil {
			fmt.Println(err)
		}
	})
	return instanceRPool
}
```


2.  生产者
```
var wg sync.WaitGroup
	for i:=0;i<100000; i++ {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			data:=rabbitmq.GetRabbitMqDataFormat("testChange5", rabbitmq.EXCHANGE_TYPE_TOPIC, "textQueue5", "/", fmt.Sprintf("这里是数据%d", num))
			_=instanceRPool.Push(data)
		}(i)
	}
	wg.Wait()
```

3. 消费者
> 可定义多个消息者事件, 不通交换机, 队列, 路由
>
> 每个事件独立
>

```
nomrl := &rabbitmq.ConsumeReceive{
#定义消费者事件
		ExchangeName: "testChange31",//队列名称
        ExchangeType: rabbitmq.EXCHANGE_TYPE_DIRECT,
        Route:        "",
        QueueName:    "testQueue31",
        IsTry:true,//是否重试
        MaxReTry: 5,//最大重试次数
        EventFail: func(code int, e error, data []byte) {
        	fmt.Printf("error:%s", e)
        },
        EventSuccess: func(data []byte)bool {//如果返回true 则无需重试
        	fmt.Printf("data:%s\n", string(data))
        	return true
        },
	}
	instanceConsumePool.RegisterConsumeReceive(nomrl)

	err := instanceConsumePool.RunConsume()
	if err != nil {
		fmt.Println(err)
	}
```

4. 错误码说明
> 错误码为
>
> 1. 生产者push时返回的  *RabbitMqError
> 2. 消费者事件监听回返的 code
>
|错误码|说明|
|-|-|
|501|生产者发送超过最大重试次数|
|502|获取信道失败, 一般为认道队列数用尽|
|503|交换机/队列/绑定失败|
|504|连接失败|
|506|信道创建失败|
|507|超过最大重试次数|
