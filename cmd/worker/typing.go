package worker

import (
	"context"
	"time"

	"github.com/nsqio/go-nsq"
)

type callbackHandler func(ctx context.Context, msg *nsq.Message) error

// TopicPolicy --- 每个 Topic 的独立策略 ---
type TopicPolicy struct {
	Topic       string
	Channel     string
	Concurrency int                        // goroutine 个数（go-nsq 自带并发处理）
	MaxInFlight int                        // 每个 consumer 的 MaxInFlight
	MaxAttempts uint16                     // 超过后进 DLQ（为 0 则使用全局/默认）
	DLQTopic    string                     // 为空时默认 "dlq.<topic>"
	Backoff     func(uint16) time.Duration // attempts -> delay（可选）
	UseLookupd  bool                       // true 用 lookupd，否则直连 nsqd
	LookupdHTTP []string                   // 覆盖 pool 级别的 lookupd 地址
	NSQd        []Nsqd                     // 覆盖 pool 级别的 nsqd 列表

	// 业务处理（收到原始 nsq.Message；你也可以在这里解码 Envelope）
	Handler callbackHandler
}

type Nsqd struct {
	IsHttp  bool   `json:"is_http"`
	Address string `json:"address"`
}
