package main

import (
	"DuckBox/Rebuilder"
	"DuckBox/conf"
	"DuckBox/controllers"
	"fmt"
	"github.com/nsqio/go-nsq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	type Route struct {
		TopicName   string `json:"topic_name"`
		ChannelName string `json:"channel_name"`
		Callback    nsq.Handler
	}

	r := []Route{
		{
			TopicName:   "pay",
			ChannelName: "default-sys",
			Callback:    nil,
		},
		{
			TopicName:   "order",
			ChannelName: "default-sys",
			Callback:    &controllers.OrderController{},
		},
		{
			TopicName:   "program",
			ChannelName: "search",
			Callback:    &controllers.HandlerProgramToSendSearch{},
		},
		{
			TopicName:   "program",
			ChannelName: "cache",
			Callback:    &controllers.ProgramCacheControllers{},
		},
		{
			TopicName:   "video",
			ChannelName: "cache",
			Callback:    &controllers.VideoCacheControllers{},
		},
		{
			TopicName:   "video",
			ChannelName: "search",
			Callback:    &controllers.HandlerVideoToSendSearch{},
		},
	}

	//rebuild video index
	go func() {
		err := Rebuilder.RebuildVideoIndex()
		if err != nil {
			return
		}
	}()
	//rebuild program index
	go func() {
		err := Rebuilder.RebuildProgramIndex()
		if err != nil {
			return
		}
	}()

	for _, route := range r {

		if route.Callback == nil {
			continue
		}
		go func(topicName, channelName string, callback nsq.Handler) {

			if callback == nil {
				return
			}
			if err := NewConsumer(fmt.Sprintf("%s:%s", conf.AppConfig.NSQ.Address, conf.AppConfig.NSQ.Port), topicName, channelName, callback); err != nil {
				//log
				controllers.Log("new-consumer-failed-message", err.Error(), controllers.LogError)
			}
		}(route.TopicName, route.ChannelName, route.Callback)
	}

	var c = make(chan os.Signal)
	//监听信号
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGUSR2)

	select {
	case <-c:

	}
}

func NewConsumer(address, topic, channel string, handler nsq.Handler) error {

	config := nsq.NewConfig()
	config.WriteTimeout = time.Second * 6

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		return err
	}
	consumer.AddHandler(handler)
	//connect
	err = consumer.ConnectToNSQLookupd(address)
	if err != nil {
		return err
	}
	//
	return nil
}
