package kafka

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/libpub/golib/logger"
	"github.com/libpub/golib/utils"
	k "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

// CallBack .回调函数
type CallBack func([]byte)

// Consumer 消费者.
type Consumer struct {
	Base
	Readers map[string]*k.Reader // 每一个topic 一个reader
	// Params     map[string]string    // 配置参数
	running    map[string]bool // 用于设置reader 是否要关闭连接
	cancels    map[string]context.CancelFunc
	Brokers    []string         // kafka 的节点
	OffsetDict map[string]int64 // 记录偏移量，避免在连接断开重连时候重复处理信息
}

// ConfigGroupID 配置group id.
func (c *Consumer) ConfigGroupID(groupID string) {
	c.Config["group.id"] = groupID
}

// ConfigMaxPollIntervalMS 配置两次拉取数据之间的最大间隔.
func (c *Consumer) ConfigMaxPollIntervalMS(interval int) {
	c.Config["max.poll.interval.ms"] = interval
}

// StopConsumer 停止消费.
func (c *Consumer) StopConsumer() {
	for k := range c.running {
		logger.Info.Printf("stop consumer %s", k)
		c.running[k] = false
		cancel := c.cancels[k]
		cancel()
	}
}

// Receive 订阅topic，处理消息.
// @title Receive
// @param topic 订阅的topic
// @param callback ,处理接收到的信息，入参是 接收到的[]byte
func (c *Consumer) Receive(topic string, callback CallBack) error {
	if _, ok := c.Readers[topic]; ok {
		return errors.New("The topic is already subscribed")
	}
	logger.Debug.Printf("group_id:%s\n", c.Config["group.id"])
	logger.Debug.Printf("%+v", c.Config)
	groupID := c.Config["group.id"].(string)
	if groupID == "" {
		groupID = topic + "-" + utils.GenUUID()
	}
	logger.Debug.Println(groupID)
	config := k.ReaderConfig{
		Brokers:        c.Brokers,
		GroupID:        groupID,
		Topic:          topic,
		MinBytes:       1,    // 1 Byte
		MaxBytes:       10e6, // 10MB
		StartOffset:    k.LastOffset,
		CommitInterval: 1 * time.Second,
		ErrorLogger:    logger.Error,
		ReadBackoffMax: 200 * time.Millisecond,
	}
	if v, ok := c.Config["heartbeat.interval.ms"]; ok {
		config.HeartbeatInterval = time.Duration(v.(int)) * time.Millisecond
	}
	if v, ok := c.Config["session.timeout.ms"]; ok {
		config.SessionTimeout = time.Duration(v.(int)) * time.Millisecond
	}
	// if v, ok := c.Config["reconnect.backoff.ms"];ok{
	// 	config.ReadBackoffMax
	// }
	if c.Config["sasl.username"] != nil && c.Config["sasl.password"] != nil {
		logger.Debug.Println("using sasl ")
		mechanism := plain.Mechanism{
			Username: c.Config["sasl.username"].(string),
			Password: c.Config["sasl.password"].(string),
		}
		dialer := &k.Dialer{
			Timeout:       10 * time.Second,
			DualStack:     true,
			SASLMechanism: mechanism,
		}
		config.Dialer = dialer

	}

	reader := k.NewReader(config)

	c.Readers[topic] = reader
	c.running[topic] = true
	c.OffsetDict[topic] = -1
	go func() {
		defer reader.Close()
		for c.running[topic] {
			ctx, cancel := context.WithCancel(context.Background())
			c.cancels[topic] = cancel
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				logger.Error.Println(err)
			}
			if m.Offset > c.OffsetDict[topic] {
				c.OffsetDict[topic] = m.Offset
				func() {
					defer func() {
						if err := recover(); err != nil {
							logger.Error.Println(err)
						}
					}()
					callback(m.Value)
				}()
			} else {
				logger.Error.Println("skipping because of offset")
			}

		}

	}()
	return nil
}

// NewConsumer 实例化返回消费者.
func NewConsumer(hosts string, groupID string) *Consumer {

	c := &Consumer{}
	c.Config = make(map[string]interface{})
	c.Readers = make(map[string]*k.Reader)
	// c.Params = make(map[string]string)
	c.running = make(map[string]bool)
	c.cancels = make(map[string]context.CancelFunc)
	c.OffsetDict = make(map[string]int64)
	c.ConfigGroupID(groupID)
	c.Brokers = strings.Split(hosts, ",")

	return c
}
