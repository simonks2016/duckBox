package producer

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewNSQPool 创建新的连接池
func NewNSQPool(addr string, workers int, opts ...PoolOption) (*NSQPool, error) {
	if workers <= 0 {
		workers = 1
	}

	client, err := NewNSQClient(addr)
	if err != nil {
		return nil, err
	}

	p := &NSQPool{
		client:    client,
		tasks:     make(chan *PublishTask, 1000), // 默认容量
		quit:      make(chan struct{}),
		backoffFn: defaultBackoff,
	}
	for _, o := range opts {
		o(p)
	}

	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go p.startWorker()
	}

	return p, nil
}

func (p *NSQPool) Close() {
	if p.closed.Swap(true) {
		return
	}
	close(p.quit) // 通知 worker 停止
	p.wg.Wait()   // 等 worker 退出
	p.client.producer.Stop()
}

// StartWorker 启动每一个工作节点
func (p *NSQPool) startWorker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			return
		case task := <-p.tasks:
			if task == nil {
				continue
			}
			err := p.execute(task)
			if task.Callback != nil {
				// 避免 callback panic 影响 worker
				func() {
					defer func() { _ = recover() }()
					task.Callback(err)
				}()
			}
		}
	}
}

func (p *NSQPool) Listen() {

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer stop()

	<-ctx.Done() // 阻塞直到收到信号

	p.Close() // 如果你实现了“排空再关”的版本，这里换成 p.Run() / p.DrainAndClose(...)
}

// 执行每一项任务
func (p *NSQPool) execute(task *PublishTask) error {
	if len(task.Msg) == 0 {
		return errors.New("empty message")
	}

	var lastErr error
	for attempt := 0; attempt <= task.Retry; attempt++ {
		// 每次尝试的超时控制
		ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
		result := make(chan error, 1)

		go func() {
			var err error
			// 延时只对单条消息有效（常见 MQ 的限制）
			if task.Delay > 0 {
				if len(task.Msg) != 1 {
					err = errors.New("deferred publish requires single message")
				} else {
					err = p.client.DeferredPublish(task.Topic, task.Delay, task.Msg[0])
				}
			} else if len(task.Msg) == 1 {
				err = p.client.Publish(task.Topic, task.Msg[0])
			} else {
				err = p.client.MultiPublish(task.Topic, task.Msg)
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
			return nil
		}
		// 失败则退避一下（最后一次不需要 sleep）
		if attempt < task.Retry {
			time.Sleep(p.backoffFn(attempt))
		}
	}
	return lastErr
}

// 提交任务
var (
	ErrPoolClosed = errors.New("nsq pool closed")
	ErrQueueFull  = errors.New("nsq pool queue is full")
)

func (p *NSQPool) submitTask(task *PublishTask) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}
	if p.trySend {
		select {
		case p.tasks <- task:
			return nil
		default:
			return ErrQueueFull
		}
	}

	if task.Delay > 0 {
		if len(task.Msg) > 1 {
			return errors.New("delay message just support the single publish")
		}
	}

	p.tasks <- task
	return nil
}

// Submit 提交任务
func (p *NSQPool) Submit(topic string, msg [][]byte, retry int, timeout time.Duration, callback func(error)) error {
	return p.submitTask(&PublishTask{
		Topic:    topic,
		Msg:      msg,
		Retry:    retry,
		Timeout:  timeout,
		Callback: callback,
	})

}

// SingleSubmit 单条快捷方法（阻塞等待结果）
func (p *NSQPool) SingleSubmit(topic string, msg []byte) error {
	done := make(chan error, 1) // 避免 callback 竞争阻塞
	_ = p.submitTask(&PublishTask{
		Topic:   topic,
		Msg:     [][]byte{msg},
		Retry:   3,
		Timeout: 2 * time.Second,
		Callback: func(err error) {
			done <- err
		},
	})
	return <-done
}

// CustomSubmit 自定义提交任务
func (p *NSQPool) CustomSubmit(options ...SubmitTaskOption) error {

	for _, opt := range options {
		task := &PublishTask{
			Topic:    opt.GetTopic(),
			Msg:      opt.GetMsg(),
			Retry:    opt.GetRetry(),
			Timeout:  opt.GetTimeout(),
			Delay:    opt.GetDelay(),
			Callback: opt.GetCallback(),
		}
		if err := p.submitTask(task); err != nil {
			return err
		}
	}
	return nil
}

// 默认退避（指数 + 抖动
func defaultBackoff(attempt int) time.Duration {
	// base 50ms，指数增长，上限 2s，加 0~100ms 抖动
	base := 50 * time.Millisecond
	max := 2 * time.Second
	d := base << attempt
	if d > max {
		d = max
	}
	jitter := time.Duration(rand.Intn(100)) * time.Millisecond
	return d + jitter
}
