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

func (c *NSQClient) Publish(topic string, msg []byte) error {
	return c.producer.Publish(topic, msg)
}

func (c *NSQClient) MultiPublish(topic string, msgs [][]byte) error {
	return c.producer.MultiPublish(topic, msgs)
}
