package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/simonks2016/duckBox/cmd/producer"
)

const (
	KeyPool = "worker.nsq_consumer_pool"
)

// --- Pool 自身配置 ---
type NSQConsumerPool struct {
	Topics               []*TopicPolicy `json:"topics"`
	LookUpDHttpAddresses []string       `json:"lookupd_http_addresses"`
	NSQd                 []Nsqd         `json:"nsqd"`

	HeartbeatSec       int `json:"heartbeat_sec"`
	LookupdPollSec     int `json:"lookupd_poll_sec"`
	DefaultMaxInFlight int `json:"default_max_in_flight"`
	DefaultMaxWorkers  int `json:"default_max_workers"`
	DefaultMaxAttempts uint16

	logger          *slog.Logger
	producerPool    *producer.NSQPool
	consumers       []*nsq.Consumer
	mu              sync.Mutex
	wg              sync.WaitGroup
	ShutdownTimeout time.Duration
	ctx             context.Context
}

// 默认值
func (p *NSQConsumerPool) Default() *NSQConsumerPool {
	if p.DefaultMaxWorkers == 0 {
		p.DefaultMaxWorkers = 15
	}
	if p.DefaultMaxInFlight == 0 {
		p.DefaultMaxInFlight = 64
	}
	if p.LookupdPollSec == 0 {
		p.LookupdPollSec = 15
	}
	if p.HeartbeatSec == 0 {
		p.HeartbeatSec = 30
	}
	if p.DefaultMaxAttempts == 0 {
		p.DefaultMaxAttempts = 5
	}
	if p.logger == nil {
		p.logger = slog.Default()
	}
	return p
}

func (p *NSQConsumerPool) AddHandler(topic ...*TopicPolicy) *NSQConsumerPool {
	p.Topics = append(p.Topics, topic...)
	return p
}

func NewNSQConsumerPool(lookupdHTTP []string, nodes ...Nsqd) *NSQConsumerPool {
	d := &NSQConsumerPool{
		LookUpDHttpAddresses: lookupdHTTP,
		NSQd:                 nodes,
	}
	return d.Default()
}

func (p *NSQConsumerPool) WithLogger(l *slog.Logger) *NSQConsumerPool {
	p.logger = l
	return p
}

// Start 启动消费者池
// 建议：在结构体里加一个关停超时
// type NSQConsumerPool struct { ... ShutdownTimeout time.Duration ... }
func (p *NSQConsumerPool) Start(ctx context.Context, producerPool *producer.NSQPool) error {
	p.mu.Lock()
	p.producerPool = producerPool
	p.ctx = ctx
	p.mu.Unlock()

	var started []*nsq.Consumer
	stopChan := make(chan struct{})

	// Handle OS signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for _, pol := range p.Topics {
		cons, err := p.newConsumer(pol)
		if err != nil {
			p.Stop(ctx)
			return fmt.Errorf("failed to create consumer for topic %s: %w", pol.Topic, err)
		}

		err = p.connectConsumer(cons, pol)
		if err != nil {
			p.Stop(ctx)
			return fmt.Errorf("failed to connect consumer for topic %s: %w", pol.Topic, err)
		}

		p.logger.Info("nsq consumer started",
			"topic", pol.Topic,
			"channel", pol.Channel,
			"concurrency", p.getConcurrency(pol),
			"max_in_flight", pol.MaxInFlight)

		// Listen on stopChan for graceful shutdown
		p.wg.Add(1)
		go func(c *nsq.Consumer) {
			defer p.wg.Done()
			<-stopChan
			c.Stop()
		}(cons)

		started = append(started, cons)
	}

	p.mu.Lock()
	p.consumers = append(p.consumers, started...)
	p.mu.Unlock()

	select {
	case <-stop:
		p.logger.Info("received OS signal, shutting down consumers...")
	case <-ctx.Done():
		p.logger.Info("context cancelled, shutting down consumers...")
	}

	close(stopChan)
	p.wg.Wait()
	p.logger.Info("all consumers shut down successfully.")
	return nil
}

// 封装创建 consumer 的逻辑，避免重复
func (p *NSQConsumerPool) newConsumer(pol *TopicPolicy) (*nsq.Consumer, error) {
	cfg := nsq.NewConfig()
	cfg.HeartbeatInterval = time.Duration(p.HeartbeatSec) * time.Second
	cfg.LookupdPollInterval = time.Duration(p.LookupdPollSec) * time.Second
	if pol.MaxInFlight > 0 {
		cfg.MaxInFlight = pol.MaxInFlight
	} else {
		cfg.MaxInFlight = p.DefaultMaxInFlight
	}

	channel := pol.Channel
	if channel == "" {
		channel = "default-workers"
	}

	cons, err := nsq.NewConsumer(pol.Topic, channel, cfg)
	if err != nil {
		return nil, err
	}

	concurrency := p.getConcurrency(pol)
	maxAttempts := p.getMaxAttempts(pol)
	dlqTopic := p.getDLQTopic(pol)

	wrapped := p.wrapHandler(*pol, maxAttempts, dlqTopic)
	cons.AddConcurrentHandlers(wrapped, concurrency)

	return cons, nil
}

// 封装连接 consumer 的逻辑，避免重复
func (p *NSQConsumerPool) connectConsumer(cons *nsq.Consumer, pol *TopicPolicy) error {
	if pol.UseLookupd {
		addrs := pol.LookupdHTTP
		if len(addrs) == 0 {
			addrs = p.LookUpDHttpAddresses
		}
		if len(addrs) == 0 {
			return fmt.Errorf("no lookupd_http_addresses found")
		}
		return cons.ConnectToNSQLookupds(addrs)
	}

	nodes := pol.NSQd
	if len(nodes) == 0 {
		nodes = p.NSQd
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no nsqd nodes found")
	}

	for _, n := range nodes {
		if err := cons.ConnectToNSQD(n.Address); err != nil {
			return err
		}
	}
	return nil
}

// 辅助函数：简化获取配置值的逻辑
func (p *NSQConsumerPool) getConcurrency(pol *TopicPolicy) int {
	if pol.Concurrency > 0 {
		return pol.Concurrency
	}
	return p.DefaultMaxWorkers
}

func (p *NSQConsumerPool) getMaxAttempts(pol *TopicPolicy) uint16 {
	if pol.MaxAttempts > 0 {
		return pol.MaxAttempts
	}
	return p.DefaultMaxAttempts
}

func (p *NSQConsumerPool) getDLQTopic(pol *TopicPolicy) string {
	if pol.DLQTopic != "" {
		return pol.DLQTopic
	}
	return "dlq." + pol.Topic
}

// Stop 关闭消费者池
func (p *NSQConsumerPool) Stop(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, c := range p.consumers {
		c.Stop()
	}
	p.consumers = nil

	if p.producerPool != nil {
		p.producerPool.Close()
	}
	p.logger.Info("nsq consumer pool stopped")

}

// wrapHandler 处理消息函数
func (p *NSQConsumerPool) wrapHandler(pol TopicPolicy, maxAttempts uint16, dlqTopic string) nsq.Handler {
	return nsq.HandlerFunc(func(m *nsq.Message) error {
		ctx := p.ctx
		ctx = WithPool(ctx, p)

		start := time.Now()
		err := pol.Handler(ctx, m)

		// 正常
		if err == nil {
			// go-nsq 会自动 FIN（因为返回 nil）
			return nil
		}

		// 分类
		var dlq DropToDLQ
		var per PermanentError
		var tr TransientError

		switch {
		case errors.As(err, &dlq):
			// 业务要求直接进 DLQ（原样 body）
			_ = p.publishDLQ(dlqTopic, m.Body)
			p.logFail(pol.Topic, m, "drop_to_dlq", dlq.Reason, start)
			return nil // ACK，不再重试

		case errors.As(err, &per):
			// 永久性错误：ACK，必要时旁路到 reject.*（此处不做）
			p.logFail(pol.Topic, m, "permanent", per.Error(), start)
			return nil

		case errors.As(err, &tr):
			// 可重试错误：看看尝试次数
			if m.Attempts >= maxAttempts {
				_ = p.publishDLQ(dlqTopic, m.Body)
				p.logFail(pol.Topic, m, "max_attempts_exceeded", tr.Error(), start)
				return nil
			}
			// backoff
			if pol.Backoff != nil {
				delay := pol.Backoff(m.Attempts)
				// 自定义退避
				m.Requeue(delay) // 自定义退避
				p.logRetry(pol.Topic, m, delay, tr.Error(), start)
				return nil
			}
			// 使用 go-nsq 默认退避：返回 error
			p.logRetry(pol.Topic, m, 0, tr.Error(), start)
			return err

		default:
			// 未分类：按可重试处理
			if m.Attempts >= maxAttempts {
				_ = p.publishDLQ(dlqTopic, m.Body)
				p.logFail(pol.Topic, m, "max_attempts_exceeded", err.Error(), start)
				return nil
			}
			p.logRetry(pol.Topic, m, 0, err.Error(), start)
			return err
		}
	})
}

func (p *NSQConsumerPool) Publish(topic string, msg []byte) error {
	return p.publishDLQ(topic, msg)
}
func (p *NSQConsumerPool) MultiPublish(items ...BatchPublishItem) error {
	for _, i := range items {
		if err := p.producerPool.SingleSubmit(i.GetTopic(), i.GetBody()); err != nil {
			return err
		}
	}
	return nil
}

// publishDLQ 发送死信
func (p *NSQConsumerPool) publishDLQ(topic string, body []byte) error {
	// 原样 body 发 DLQ
	if p.producerPool == nil {
		return fmt.Errorf("producer nil, cannot publish DLQ")
	}
	return p.producerPool.SingleSubmit(topic, body)
}

// logRetry 记录重试
func (p *NSQConsumerPool) logRetry(topic string, m *nsq.Message, delay time.Duration, reason string, start time.Time) {
	p.logger.Warn("nsq retry",
		"topic", topic,
		"msg_id", m.ID,
		"attempts", m.Attempts,
		"delay", delay.String(),
		"reason", reason,
		"latency_ms", time.Since(start).Milliseconds(),
	)
}

// logFail 记录错误
func (p *NSQConsumerPool) logFail(topic string, m *nsq.Message, kind string, reason string, start time.Time) {
	p.logger.Error("nsq failed",
		"topic", topic,
		"msg_id", m.ID,
		"attempts", m.Attempts,
		"kind", kind,
		"reason", reason,
		"latency_ms", time.Since(start).Milliseconds(),
	)
}

func WithPool(ctx context.Context, pool *NSQConsumerPool) context.Context {
	return context.WithValue(ctx, KeyPool, pool)
}

func GetPool(ctx context.Context) *NSQConsumerPool {
	if v := ctx.Value(KeyPool); v != nil {
		if p, ok := v.(*NSQConsumerPool); ok {
			return p
		}
	}
	return nil
}

type BatchPublishItem interface {
	GetTopic() string
	GetBody() []byte
}

type d struct {
	topic string
	body  []byte
}

func (d1 *d) GetTopic() string { return d1.topic }
func (d1 *d) GetBody() []byte  { return d1.body }

func NewPublishItem(topic string, body []byte) BatchPublishItem {
	return &d{
		topic: topic,
		body:  body,
	}
}
