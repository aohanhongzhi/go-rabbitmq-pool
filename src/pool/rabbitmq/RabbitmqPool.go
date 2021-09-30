package rabbitmq

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_MAX_CONNECTION      = 5   //rabbitmq tcp 最大连接数
	DEFAULT_MAX_CONSUME_CHANNEL = 100 //最大消费channel数(一般指消费者)
	DEFAULT_MAX_CONSUME_RETRY   = 5   //消费者断线重连最大次数
	DEFAULT_PUSH_MAX_TIME       = 5   //最大重发次数

	//轮循-连接池负载算法
	LOAD_BALANCE_ROUND = 1
)

const (
	RABBITMQ_TYPE_PUBLISH = 1 //生产者
	RABBITMQ_TYPE_CONSUME = 2 //消费者
)

const (
	EXCHANGE_TYPE_FANOUT = "fanout" //  Fanout：广播，将消息交给所有绑定到交换机的队列
	EXCHANGE_TYPE_DIRECT = "direct" //Direct：定向，把消息交给符合指定routing key 的队列
	EXCHANGE_TYPE_TOPIC  = "topic"  //Topic：通配符，把消息交给符合routing pattern（路由模式） 的队列
)

/**
错误码
*/
const (
	RCODE_PUSH_MAX_ERROR                    = 501 //发送超过最大重试次数
	RCODE_GET_CHANNEL_ERROR                 = 502 //获取信道失败
	RCODE_CHANNEL_QUEUE_EXCHANGE_BIND_ERROR = 503 //交换机/队列/绑定失败
	RCODE_CONNECTION_ERROR                  = 504 //连接失败
	RCODE_PUSH_ERROR                        = 505 //消息推送失败
	RCODE_CHANNEL_CREATE_ERROR              = 506 //信道创建失败

)

/**
错误返回
*/
type RabbitMqError struct {
	Code    int
	Message string
	Detail  string
}

func (e RabbitMqError) Error() string {
	return fmt.Sprintf("Exception (%d) Reason: %q", e.Code, e.Message)
}

func NewRabbitMqError(code int, message string, detail string) *RabbitMqError {
	return &RabbitMqError{Code: code, Message: message, Detail: detail}
}

/**
消费者注册接收数据
*/
type ConsumeReceive struct {
	ExchangeName   string           //交换机
	ExchangeType   string           //交换机类型
	Route          string           //路由
	QueueName      string           //队列名称
	EventSuccess   func([]byte)     //成功事件回调
	EventFail      func(int, error) //失败回调
	DeadExpireTime int              //死信队列超时时间, 如果设置此时间创建的默认为死信队列(毫秒)
}

/**
单个rabbitmq channel
*/
type rChannel struct {
	ch    *amqp.Channel
	index int32
}

type rConn struct {
	conn  *amqp.Connection
	index int32
}

type RabbitPool struct {
	maxConnection int32 // 最大连接数量
	pushMaxTime   int   //最大重发次数

	connectionIndex   int32 //记录当前使用的连接
	connectionBalance int   //连接池负载算法

	channelPool map[int64]*rChannel //channel信道池
	connections map[int][]*rConn    // rabbitmq连接池

	channelLock    sync.RWMutex //信道池锁
	connectionLock sync.Mutex   //连接锁

	rabbitLoadBalance *RabbitLoadBalance //连接池负载模式(生产者)

	consumeMaxChannel   int32             //消费者最大信道数一般指消费者
	consumeReceive      []*ConsumeReceive //消费者注册事件
	consumeMaxRetry     int32             //消费者断线重连最大次数
	consumeCurrentRetry int32             //当前重连次数
	pushCurrentRetry    int32             //当前推送重连交数

	clientType int //客户端类型 生产者或消费者 默认为生产者

	errorChanel chan *amqp.Error //错误捕捉channel

	connectStatus bool

	host     string //服务ip
	port     int    //服务端口
	user     string //用户名
	password string //密码
}

/**
初始化生产者
*/
func NewProductPool() *RabbitPool {
	return newRabbitPool(RABBITMQ_TYPE_PUBLISH)
}

/**
初始化消费者
*/
func NewConsumePool() *RabbitPool {
	return newRabbitPool(RABBITMQ_TYPE_CONSUME)
}

func newRabbitPool(clientType int) *RabbitPool {
	return &RabbitPool{
		clientType:          clientType,
		consumeMaxChannel:   DEFAULT_MAX_CONSUME_CHANNEL,
		maxConnection:       DEFAULT_MAX_CONNECTION,
		pushMaxTime:         DEFAULT_PUSH_MAX_TIME,
		connectionBalance:   LOAD_BALANCE_ROUND,
		connectionIndex:     0,
		consumeMaxRetry:     DEFAULT_MAX_CONSUME_RETRY,
		consumeCurrentRetry: 0,
		pushCurrentRetry:    0,
		connectStatus:       false,
		connections:         make(map[int][]*rConn, 2),
		channelPool:         make(map[int64]*rChannel, 1),
		rabbitLoadBalance:   NewRabbitLoadBalance(),
		errorChanel:         make(chan *amqp.Error),
	}
}

/**
设置消费者最大信道数
*/
func (r *RabbitPool) SetMaxConsumeChannel(maxConsume int32) {
	r.consumeMaxChannel = maxConsume
}

/**
设置最大连接数
*/
func (r *RabbitPool) SetMaxConnection(maxConnection int32) {
	r.maxConnection = maxConnection
}

/**
设置连接池负载算法
默认轮循
*/
func (r *RabbitPool) SetConnectionBalance(balance int) {
	r.connectionBalance = balance
}

/**
连接rabbitmq
@param host string 服务器地址
@param port int 服务端口
@param user string 用户名
@param password 密码
*/
func (r *RabbitPool) Connect(host string, port int, user string, password string) error {
	r.host = host
	r.port = port
	r.user = user
	r.password = password
	return r.initConnections(false)
}

/**
注册消费接收
*/
func (r *RabbitPool) RegisterConsumeReceive(consumeReceive *ConsumeReceive) {
	if consumeReceive != nil {
		r.consumeReceive = append(r.consumeReceive, consumeReceive)
	}
}

/**
消费者
*/
func (r *RabbitPool) RunConsume() error {
	r.clientType = RABBITMQ_TYPE_CONSUME
	if len(r.consumeReceive) == 0 {
		return errors.New("未注册消费者事件")
	}
	rConsume(r)
	return nil
}

/**
发送消息
*/
func (r *RabbitPool) Push(data *RabbitMqData) *RabbitMqError {
	return rPush(r, data, 1)
}

/**
获取当前连接
1.这里可以做负载算法, 默认使用轮循
*/
func (r *RabbitPool) getConnection() *rConn {
	changeConnectionIndex := r.connectionIndex
	currentIndex := r.rabbitLoadBalance.RoundRobin(changeConnectionIndex, r.maxConnection)
	currentNum := currentIndex - changeConnectionIndex
	atomic.AddInt32(&r.connectionIndex, currentNum)
	return r.connections[r.clientType][r.connectionIndex]
}

/**
获取信道
1.如果当前信道池不存在则创建
2.如果信息池存在则直接获取
3.每个连接池中连接维护一组信道
@param channelName string 信息道名称
*/
func (r *RabbitPool) getChannelQueue(conn *rConn, exChangeName string, exChangeType string, queueName string, route string) (*rChannel, error) {
	channelHashCode := channelHashCode(r.clientType, conn.index, exChangeName, exChangeType, queueName, route)
	if channelQueues, ok := r.channelPool[channelHashCode]; ok {
		return channelQueues, nil
	} else { //如果不存在则创建信道池
		//初始化channel
		rChannel, err := r.initChannels(conn, exChangeName, exChangeType, queueName, route)
		if err != nil {
			return nil, err
		}
		channel, err := rDeclare(conn, r.clientType, rChannel, exChangeName, exChangeType, queueName, route, false, 0)
		if err != nil {
			return nil, err
		}
		rChannel.ch = channel.ch
		r.channelPool[channelHashCode] = rChannel
		return rChannel, nil
	}
}

/**
初始化连接池
*/
func (r *RabbitPool) initConnections(isLock bool) error {
	r.connections[r.clientType] = []*rConn{}
	var i int32 = 0
	for i = 0; i < r.maxConnection; i++ {
		itemConnection, err := rConnect(r, isLock)
		if err != nil {
			return err
		} else {
			r.connections[r.clientType] = append(r.connections[r.clientType], &rConn{conn: itemConnection, index: i})
		}
	}
	return nil
}

/**
初始化信道池
*/
func (r *RabbitPool) initChannels(conn *rConn, exChangeName string, exChangeType string, queueName string, route string) (*rChannel, error) {
	//var i int32 = 0
	//var channelQueue = NewChannelQueue()
	//for i = 0; i < r.maxChannel; i++ {
	//	channel, err := rCreateChannel(conn)
	//	if err != nil {
	//		return nil, err
	//	}
	//	channelQueue.Add(&rChannel{ch: channel, index: i})
	//}
	channel, err := rCreateChannel(conn)
	if err != nil {
		return nil, err
	}
	rChannel := &rChannel{ch: channel, index: 0}
	return rChannel, nil
}

/**
原rabbitmq连接
*/
func rConnect(r *RabbitPool, islock bool) (*amqp.Connection, error) {
	connectionUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/", r.user, r.password, r.host, r.port)
	client, err := amqp.Dial(connectionUrl)
	if err != nil {
		return nil, err
	}
	return client, nil
}

/**
创建rabbitmq信道
*/
func rCreateChannel(conn *rConn) (*amqp.Channel, error) {
	ch, err := conn.conn.Channel()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Create Connect Channel Error: %s", err.Error()))
	}
	return ch, nil
}

/**
绑定并声明
@param rconn *rConn tcp连接对象
@param clientType int 客户端类型
@param channel 信道
@param exChangeName 交换机名称
@param exChangeType 交换机类型
@param queueName 队列名称
@param route 路由key
@param isDeadQueue 是否是死信队列
@param deadQueueExpireTime int 死信队列到期时间
*/
func rDeclare(rconn *rConn, clientType int, channel *rChannel, exChangeName string, exChangeType string, queueName string, route string, isDeadQueue bool, deadQueueExpireTime int) (*rChannel, error) {
	if clientType == RABBITMQ_TYPE_PUBLISH {
		if (len(exChangeType) == 0) || (exChangeType != EXCHANGE_TYPE_DIRECT && exChangeType != EXCHANGE_TYPE_FANOUT && exChangeType != EXCHANGE_TYPE_TOPIC) {
			return channel, errors.New("交换机类型错误")
		}
	}
	newChannel := channel.ch
	//if clientType == RABBITMQ_TYPE_PUBLISH {
	//	err := newChannel.ExchangeDeclarePassive(exChangeName, exChangeType, true, false, false, false, nil)
	//	if err != nil {
	// 注册交换机
	// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
	// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
	err := newChannel.ExchangeDeclare(exChangeName, exChangeType, true, false, false, false, nil)
	if err != nil {
		fmt.Println("xxx2", err)
		return nil, errors.New(fmt.Sprintf("MQ注册交换机失败:%s", err))
	}
	//}
	//}

	// 用于检查队列是否存在,已经存在不需要重复声明
	//queue, err := newChannel.QueueDeclarePassive(
	//	queueName,
	//	true,  // durable  设置是否持久化, true表示队列为持久化, 持久化的队列会存盘, 在服务器重启的时候会保证不丢失相关信息
	//	false, // delete when unused 不使用时自动删除
	//	false, // exclusive  是否排他
	//	false, // no-wait 是否等待服务器返回
	//	nil,   // arguments 相关参数，目前一般为nil
	//)
	//if err != nil {
	// 队列不存在,声明队列
	// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
	// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他

	if (clientType != RABBITMQ_TYPE_PUBLISH && exChangeType != EXCHANGE_TYPE_FANOUT) || (clientType == RABBITMQ_TYPE_CONSUME && (exChangeType == EXCHANGE_TYPE_FANOUT || exChangeType == EXCHANGE_TYPE_DIRECT)) {
		deadArgMap := make(map[string]interface{})
		if isDeadQueue {
			deadArgMap["x-message-ttl"] = deadQueueExpireTime
			deadArgMap["x-dead-letter-exchange"] = exChangeName
			deadArgMap["x-dead-letter-routing-key"] = route
		}
		queue, err := newChannel.QueueDeclare(queueName, true, false, false, true, deadArgMap)
		if err != nil {
			//fmt.Println(err)
			return nil, errors.New(fmt.Sprintf("MQ注册队列失败:%s", err))
		}
		//}

		err = newChannel.QueueBind(queue.Name, route, exChangeName, false, nil)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("MQ绑定队列失败:%s", err))
		}
	}
	channel.ch = newChannel
	return channel, nil
}

/**
消费者处理
*/
func rConsume(pool *RabbitPool) {
	for _, v := range pool.consumeReceive {
		go func(pool *RabbitPool, receive *ConsumeReceive) {
			rListenerConsume(pool, receive)
		}(pool, v)
	}
	/**
	创建一个协程监听任务
	*/
	select {
	//case data := <-pool.errorChanel:
	case <-pool.errorChanel:
		statusLock.Lock()
		status = true
		statusLock.Unlock()
		retryConsume(pool)
	}

}

/**
重连处理
*/
func retryConsume(pool *RabbitPool) {
	fmt.Printf("2秒后开始重试:[%d]\n", pool.consumeCurrentRetry)
	atomic.AddInt32(&pool.consumeCurrentRetry, 1)
	time.Sleep(time.Second * 2)
	_, err := rConnect(pool, true)
	if err != nil {
		retryConsume(pool)
	} else {
		statusLock.Lock()
		status = false
		statusLock.Unlock()
		_ = pool.initConnections(false)
		rConsume(pool)
	}

}

/**
监听消费
*/
func rListenerConsume(pool *RabbitPool, receive *ConsumeReceive) {
	var i int32 = 0
	for i = 0; i < pool.consumeMaxChannel; i++ {
		go func(num int32, p *RabbitPool, r *ConsumeReceive) {
			consumeTask(num, p, r)
		}(i, pool, receive)
	}
}

var statusLock sync.Mutex
var status bool = false

func setConnectError(pool *RabbitPool, code int, message string) {
	statusLock.Lock()
	defer statusLock.Unlock()

	if !status {
		pool.errorChanel <- &amqp.Error{
			Code:   code,
			Reason: message,
		}
	}
	status = true
}

/***
消费任务
*/

func consumeTask(num int32, pool *RabbitPool, receive *ConsumeReceive) {
	//获取请求连接
	closeFlag := false
	pool.connectionLock.Lock()
	conn := pool.getConnection()
	pool.connectionLock.Unlock()
	//生成处理channel 根据最大channel数处理
	channel, err := rCreateChannel(conn)
	if err != nil {
		if receive.EventFail != nil {
			receive.EventFail(RCODE_CHANNEL_CREATE_ERROR, NewRabbitMqError(RCODE_CHANNEL_CREATE_ERROR, "channel create error", err.Error()))
		}
		return
	}
	defer func() {
		_ = channel.Close()
		_ = conn.conn.Close()
	}()
	//defer
	closeChan := make(chan *amqp.Error, 1)
	rChanels := &rChannel{ch: channel, index: num}
	//申请并绑定
	isDeadQueue :=false
	if receive.DeadExpireTime > 0{
		isDeadQueue = true
	}
	rChanels, err = rDeclare(conn, pool.clientType, rChanels, receive.ExchangeName, receive.ExchangeType, receive.QueueName, receive.Route, isDeadQueue, receive.DeadExpireTime)
	if err != nil {
		if receive.EventFail != nil {
			receive.EventFail(RCODE_CHANNEL_QUEUE_EXCHANGE_BIND_ERROR, NewRabbitMqError(RCODE_CHANNEL_QUEUE_EXCHANGE_BIND_ERROR, "交换机/队列/绑定失败", err.Error()))
		}
		return
	}
	// 获取消费通道
	//确保rabbitmq会一个一个发消息
	_ = channel.Qos(1, 0, false)
	msgs, err := channel.Consume(
		receive.QueueName, // queue
		"",                // consumer
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if nil != err {
		if receive.EventFail != nil {
			receive.EventFail(RCODE_GET_CHANNEL_ERROR, NewRabbitMqError(RCODE_GET_CHANNEL_ERROR, fmt.Sprintf("获取队列 %s 的消费通道失败", receive.QueueName), err.Error()))
		}
		return
	}

	//一旦消费者的channel有错误，产生一个amqp.Error，channel监听并捕捉到这个错误
	notifyClose := channel.NotifyClose(closeChan)

	for {
		select {
		case data := <-msgs:
			_ = data.Ack(false)
			if receive.EventSuccess != nil {
				receive.EventSuccess(data.Body)
			}
		//一但有错误直接返回 并关闭信道
		case e := <-notifyClose:
			if receive.EventFail != nil {
				receive.EventFail(RCODE_CONNECTION_ERROR, NewRabbitMqError(RCODE_CONNECTION_ERROR, fmt.Sprintf("消息处理中断: queue:%s\n", receive.QueueName), e.Error()))
			}
			setConnectError(pool, e.Code, fmt.Sprintf("消息处理中断: %s", e.Error()))
			closeFlag = true
		}
		if closeFlag {
			break
		}
	}
}

/**
发送消息
*/
func rPush(pool *RabbitPool, data *RabbitMqData, sendTime int) *RabbitMqError {
	if sendTime >= pool.pushMaxTime {
		return NewRabbitMqError(RCODE_PUSH_MAX_ERROR, "重试超过最大次数", "")
	}
	pool.channelLock.Lock()
	conn := pool.getConnection()
	rChannel, err := pool.getChannelQueue(conn, data.ExchangeName, data.ExchangeType, data.QueueName, data.Route)
	pool.channelLock.Unlock()
	if err != nil {
		return NewRabbitMqError(RCODE_GET_CHANNEL_ERROR, "获取信道失败", err.Error())
	} else {
		err = rChannel.ch.Publish(data.ExchangeName, data.Route, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(data.Data),
		})
		if err != nil { //如果消息发送失败, 重试发送
			//pool.channelLock.Unlock()
			//如果没有发送成功,休息两秒重发
			time.Sleep(time.Second * 2)
			sendTime++
			return rPush(pool, data, sendTime)
		}

	}
	return nil
}

/**
信道hashcode
*/
func channelHashCode(clientType int, connIndex int32, exChangeName string, exChangeType string, queueName string, route string) int64 {
	channelHashCode := hashCode(fmt.Sprintf("%d-%d-%s-%s-%s-%s", clientType, connIndex, exChangeName, exChangeType, queueName, route))
	return channelHashCode
}

/**
计算hashcode唯一值
*/
func hashCode(s string) int64 {
	v := int64(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return -1
}