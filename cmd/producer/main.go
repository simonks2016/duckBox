package producer

import (
	"time"

	"github.com/nsqio/go-nsq"
)

type NSQClient struct {
	producer *nsq.Producer
	addr     string
}

func NewNSQClient(addr string) (*NSQClient, error) {
	config := nsq.NewConfig()
	config.DialTimeout = 10 * time.Second
	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil, err
	}
	return &NSQClient{producer: producer, addr: addr}, nil
}

// Publish 发送单个消息
func (c *NSQClient) Publish(topic string, msg []byte) error {
	return c.producer.Publish(topic, msg)
}

// MultiPublish 同一个Topic多次批量发送消息
func (c *NSQClient) MultiPublish(topic string, msgs [][]byte) error {
	return c.producer.MultiPublish(topic, msgs)
}

// Close 发送延时消息
func (c *NSQClient) Close() {
	c.producer.Stop()
}

// DeferredPublish 延时消息发送
func (c *NSQClient) DeferredPublish(topic string, delay time.Duration, msg []byte) error {
	return c.producer.DeferredPublish(topic, delay, msg)
}
