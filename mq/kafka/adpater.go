package kafka

import (
	"fmt"
	"time"

	"github.com/libpub/golib/mq/mqenv"
	"github.com/libpub/golib/utils"
)

// kafkaInstances kafka 实例.
var kafkaInstances = map[string]*KafkaWorker{}

// Config kafkav2 配置参数.
type Config struct {
	Hosts             string
	Partition         int
	PrivateTopic      string
	GroupID           string
	MaxPollIntervalMS int
	// 消息类型:
	//direct:组播,订阅同一个topic，消费者组会相同，一条消息只会被组内一个消费者接收
	//fanout:广播,订阅同一个topic，但是消费者组会使用uuid，所有组都会收到信息
	MessageType string `yaml:"messageType" json:"messageType"`
	// kerberos 认证需要配置
	KerberosServiceName string
	KerberosKeytab      string
	KerberosPrincipal   string
	// plain 认证需要配置
	SaslMechanisms     string
	SaslUsername       string
	SaslPassword       string
	UseOriginalContent bool `yaml:"useOriginalContent" json:"useOriginalContent"`
}

// InstStats .
type InstStats struct {
	Bytes         int64  `json:"bytes"`
	Dials         int64  `json:"connections"`
	Topic         string `json:"topic"`
	Messages      int64  `json:"messages"`
	Rebalances    int64  `json:"rebalances"`
	Errors        int64  `json:"errors"`
	Timeouts      int64  `json:"timeouts"`
	ClientID      string `json:"clientID"`
	QueueLength   int64  `json:"queueLength"`
	QueueCapacity int64  `json:"queueCapacity"`
}

// Stats struct
type Stats struct {
	Consumer InstStats `json:"consumer"`
	Producer InstStats `json:"producer"`
}

// InitKafka 初始化kafka.
func InitKafka(mqConnName string, config Config) (*KafkaWorker, error) {
	instance, ok := kafkaInstances[mqConnName]
	if !ok {
		if config.PrivateTopic == "" {
			config.PrivateTopic = "rpc-" + utils.GenUUID()
		} else {
			config.PrivateTopic += "-" + utils.GenUUID()[:10]
		}
		if config.MessageType == "fanout" {
			config.GroupID = utils.GenUUID()
			config.PrivateTopic = ""
		}
		instance = NewKafkaWorker(config.Hosts, config.Partition, config.PrivateTopic, config.GroupID)
		instance.UseOriginalContent = config.UseOriginalContent
		// if config.KerberosServiceName != "" && config.KerberosKeytab != "" && config.KerberosPrincipal != "" {
		// 	instance.Producer.ConfigKerberosServiceName(config.KerberosServiceName)
		// 	instance.Producer.ConfigKerberosKeyTab(config.KerberosKeytab)
		// 	instance.Producer.ConfigKerberosPrincipal(config.KerberosPrincipal)
		// 	instance.Producer.ConfigSecurityProtocol("sasl_plaintext")

		// 	instance.Consumer.ConfigKerberosServiceName(config.KerberosServiceName)
		// 	instance.Consumer.ConfigKerberosKeyTab(config.KerberosKeytab)
		// 	instance.Consumer.ConfigKerberosPrincipal(config.KerberosPrincipal)
		// 	instance.Consumer.ConfigSecurityProtocol("sasl_plaintext")
		// }
		if config.SaslUsername != "" && config.SaslPassword != "" {
			instance.Producer.ConfigSaslUserName(config.SaslUsername)
			instance.Producer.ConfigSaslPassword(config.SaslPassword)
			instance.Producer.ConfigSecurityProtocol("sasl_plaintext")

			instance.Consumer.ConfigSaslUserName(config.SaslUsername)
			instance.Consumer.ConfigSaslPassword(config.SaslPassword)
			instance.Consumer.ConfigSecurityProtocol("sasl_plaintext")
		}
		if config.MaxPollIntervalMS > 0 {
			instance.Consumer.ConfigMaxPollIntervalMS(config.MaxPollIntervalMS)
		}
		kafkaInstances[mqConnName] = instance
		return instance, nil
	}
	return instance, nil

}

// GetKafka 获取kafka.
func GetKafka(mqConnName string) (*KafkaWorker, error) {
	instance, ok := kafkaInstances[mqConnName]
	if ok {
		return instance, nil
	}
	return nil, fmt.Errorf("Kafka instance by %s not found", mqConnName)
}

// 停止kafka
func StopKafka(mqConnName string) error {
	instance, ok := kafkaInstances[mqConnName]
	if ok {
		instance.Consumer.StopConsumer()
		delete(kafkaInstances, mqConnName)
	}
	return fmt.Errorf("Kafka instance by %s not found", mqConnName)

}

// ConvertKafkaPacketToMQConsumerMessage 把接收到的kafkaPacket 数据转换成MQConsumerMessage.
func ConvertKafkaPacketToMQConsumerMessage(packet *KafkaPacket) mqenv.MQConsumerMessage {
	consumerMessage := mqenv.MQConsumerMessage{
		Driver:        mqenv.DriverTypeKafka,
		Queue:         packet.SendTo,
		CorrelationID: packet.CorrelationId,
		ConsumerTag:   packet.ConsumerTag,
		ReplyTo:       packet.ReplyTo,
		MessageID:     packet.MessageId,
		AppID:         packet.AppId,
		UserID:        packet.UserId,
		ContentType:   packet.ContentType,
		Exchange:      packet.Exchange,
		RoutingKey:    packet.RoutingKey,
		Timestamp:     time.Unix(int64(packet.Timestamp), 0),
		Body:          packet.Body,
		Headers:       map[string]string{},
		BindData:      &packet,
	}
	if nil != packet.Headers {
		for _, h := range packet.Headers {
			consumerMessage.Headers[h.Name] = h.Value
		}
	}

	return consumerMessage
}
