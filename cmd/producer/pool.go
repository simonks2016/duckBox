package producer

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// NewNSQPool 新建NSQ连接池
func NewNSQPool(addr string, workers int) (*NSQPool, error) {
	client, err := NewNSQClient(addr)
	if err != nil {
		return nil, err
	}
	pool := &NSQPool{
		client: client,
		tasks:  make(chan *PublishTask, 1000),
		quit:   make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		go pool.Run()
	}
	return pool, nil
}

// Run 启动连接池
func (p *NSQPool) Run() {
	for {
		select {
		case task := <-p.tasks:
			var lastErr error
			for attempt := 0; attempt <= task.Retry; attempt++ {
				ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
				result := make(chan error, 1)

				go func() {
					var err error
					if len(task.Msg) == 1 {
						err = p.client.Publish(task.Topic, task.Msg[0])
					} else if len(task.Msg) > 1 {
						err = p.client.MultiPublish(task.Topic, task.Msg)
					} else {
						err = errors.New("empty message")
					}
					result <- err
				}()

				select {
				case <-ctx.Done():
					lastErr = fmt.Errorf("publish timeout: %w", ctx.Err())
				case err := <-result:
					lastErr = err
				}
				cancel()

				if lastErr == nil {
					break // 成功了就跳出 retry 循环
				}
			}
			if task.Callback != nil {
				task.Callback(lastErr)
			}
		case <-p.quit:
			return
		}
	}
}

// Submit 提交任务
func (p *NSQPool) Submit(topic string, msg [][]byte, retry int, timeout time.Duration, callback func(error)) {
	p.tasks <- &PublishTask{
		Topic:    topic,
		Msg:      msg,
		Retry:    retry,
		Timeout:  timeout,
		Callback: callback,
	}
}

// SingleSubmit 提交单个任务
func (p *NSQPool) SingleSubmit(topic string, msg ...[]byte) error {
	var done = make(chan error)
	p.tasks <- &PublishTask{
		Topic:   topic,
		Msg:     msg,
		Retry:   3,
		Timeout: time.Second * 2,
		Callback: func(err error) {
			done <- err
		},
	}
	return <-done
}

func (p *NSQPool) Close() {
	close(p.quit)
	p.client.producer.Stop()
}
