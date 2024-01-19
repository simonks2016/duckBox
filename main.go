package main

import (
	"DuckBox/Define"
	"DuckBox/Rebuilder"
	"DuckBox/conf"
	"DuckBox/controllers"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/panjf2000/ants/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {

	type Route struct {
		TopicName   string `json:"topic_name"`
		ChannelName string `json:"channel_name"`
		Callback    nsq.Handler
	}

	var r []*Route

	r = append(r, &Route{
		TopicName:   Define.ProgramTopic,
		ChannelName: "gorse",
		Callback:    &controllers.SendProgramToGorse{},
	}, &Route{
		TopicName:   Define.VideoTopic,
		ChannelName: "gorse",
		Callback:    &controllers.SendVideoToGorse{},
	}, &Route{
		TopicName:   Define.CustomerTopic,
		ChannelName: "gorse",
		Callback:    &controllers.SendUserToGorse{},
	}, &Route{
		TopicName:   Define.LikeTopic,
		ChannelName: "gorse",
		Callback:    &controllers.SendLike2Gorse{},
	}, &Route{
		TopicName:   Define.FollowTopic,
		ChannelName: "gorse",
		Callback:    &controllers.AfterFollow{},
	}, &Route{
		TopicName:   Define.RecordTopic,
		ChannelName: "gorse",
		Callback:    &controllers.SendFeedback2Gorse{},
	}, &Route{
		TopicName:   Define.OrderTopic,
		ChannelName: "default-sys",
		Callback:    &controllers.OrderController{},
	},
		&Route{
			TopicName:   Define.SubscribeProgramTopic,
			ChannelName: "gorse",
			Callback:    &controllers.SubscribeProgramToGorse{},
		},
		&Route{
			TopicName:   Define.ProgramTopic,
			ChannelName: "search",
			Callback:    &controllers.HandlerProgramToSendSearch{},
		},
		&Route{
			TopicName:   Define.ProgramTopic,
			ChannelName: "cache",
			Callback:    &controllers.ProgramCacheControllers{},
		},
		&Route{
			TopicName:   Define.VideoTopic,
			ChannelName: "cache",
			Callback:    &controllers.VideoCacheControllers{},
		},
		&Route{
			TopicName:   Define.VideoTopic,
			ChannelName: "search",
			Callback:    &controllers.HandlerVideoToSendSearch{},
		},
		&Route{
			TopicName:   Define.EpisodesTopic,
			ChannelName: "cache",
			Callback:    &controllers.EpisodesCacheControllers{},
		})

	var wg sync.WaitGroup
	var wgNum = 0
	var pool, _ = ants.NewPool(1000, ants.WithMaxBlockingTasks(500))
	defer func() {
		pool.Release()
	}()

	var builder = Rebuilder.Builder{SubmitTask: func(f func(*sync.WaitGroup)) {
		//set the number
		//wgNum = wgNum + 1
		//add
		wg.Add(1)
		//submit task
		err := pool.Submit(func() {
			f(&wg)
		})
		if err != nil {
			//record the error message
			controllers.Log("Received error message from thread", err.Error(), controllers.LogError)
			return
		}
	}}

	taskList := []func() error{
		builder.BuildProgramIndex,
		builder.BuildVideoIndex,
		builder.BuildRecommendItems,
		builder.BuildRecommendCustomerItem,
		builder.BuildFeedback,
	}
	//base on the route map to make task list

	for _, ro := range r {
		//if the callback is empty
		if ro.Callback == nil {
			continue
		}

		topicName := ro.TopicName
		channelName := ro.ChannelName
		callback := ro.Callback

		taskList = append(taskList, func() error {
			return NewConsumer(conf.AppConfig.NSQ.ToHost(),
				topicName,
				channelName,
				callback)
		})
	}
	//Set votes to N
	wgNum = len(taskList)
	//add the counter
	wg.Add(wgNum)
	//loop to set
	for _, f := range taskList {
		//Copy the elements in the array to the callback parameter
		var callback = f
		//submit task
		err := pool.Submit(func() {
			//build program index
			if err := callback(); err != nil {
				//record teh error message
				controllers.Log("submit task", err.Error(), controllers.LogError)
				return
			}
			//done the thread
			defer wg.Done()
		})
		if err != nil {
			//record the error message
			controllers.Log("the pool submit failed",
				err.Error(), controllers.LogError)
			return
		}
	}

	//running the thread pool
	//pool.Running()
	//wait for
	wg.Wait()
}

func NewConsumer(address, topic, channel string, handler nsq.Handler) error {

	config := nsq.NewConfig()
	config.WriteTimeout = time.Second * 6

	consumer, err := nsq.NewConsumer(topic, channel, config)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	consumer.AddHandler(handler)
	//connect
	err = consumer.ConnectToNSQLookupd(address)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	consumer.Stop()
	//
	return nil
}
