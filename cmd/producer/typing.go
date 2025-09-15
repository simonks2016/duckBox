package producer

import (
	"sync"
	"time"
)

// PublishTask 任务结构
type PublishTask struct {
	Topic    string
	Msg      [][]byte
	Callback func(error)
	Retry    int           // 最大重试次数
	Timeout  time.Duration // 每次尝试的超时时间
}

type NSQPool struct {
	client *NSQClient
	tasks  chan *PublishTask
	wg     sync.WaitGroup
	quit   chan struct{}
}
