package producer

import (
	"sync"
	"sync/atomic"
	"time"
)

type NSQPool struct {
	client           *NSQClient
	tasks            chan *PublishTask
	quit             chan struct{}
	wg               sync.WaitGroup
	closed           atomic.Bool
	trySend          bool // true: Submit 会用非阻塞入队
	backoffFn        func(int) time.Duration
	shutdownDeadline time.Duration
}

// PoolOption 构造
type PoolOption func(*NSQPool)

func WithQueueSize(n int) PoolOption {
	return func(p *NSQPool) { p.tasks = make(chan *PublishTask, n) }
}
func WithTrySend(enable bool) PoolOption {
	return func(p *NSQPool) { p.trySend = enable }
}
func WithBackoff(fn func(attempt int) time.Duration) PoolOption {
	return func(p *NSQPool) { p.backoffFn = fn }
}
