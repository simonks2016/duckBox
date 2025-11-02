package producer

import "time"

// PublishTask 任务结构
type PublishTask struct {
	Topic    string
	Msg      [][]byte      // 允许单条或多条
	Retry    int           // >=0；重试次数
	Timeout  time.Duration // 每次尝试的超时
	Delay    time.Duration // >0 则走延时发送（仅单条消息有意义）
	Callback func(error)
}

// SubmitTaskOption 提交任务配置项
type SubmitTaskOption interface {
	GetTopic() string
	GetMsg() [][]byte
	GetRetry() int
	GetTimeout() time.Duration
	GetDelay() time.Duration // 替代 IsDeferred，语义更清晰
	GetCallback() func(error)
}

func NewSingleTaskOption(topic string, msg ...[]byte) *SingleTaskOption {

	return &SingleTaskOption{
		topic: topic,
		msg:   msg,
	}
}

type SingleTaskOption struct {
	topic    string
	msg      [][]byte
	retry    int
	timeout  time.Duration
	delay    time.Duration
	callback func(error)
}

func (o *SingleTaskOption) GetTopic() string          { return o.topic }
func (o *SingleTaskOption) GetMsg() [][]byte          { return o.msg }
func (o *SingleTaskOption) GetRetry() int             { return o.retry }
func (o *SingleTaskOption) GetTimeout() time.Duration { return o.timeout }
func (o *SingleTaskOption) GetDelay() time.Duration   { return o.delay }
func (o *SingleTaskOption) GetCallback() func(error)  { return o.callback }

func (o *SingleTaskOption) WithTopic(topic string) *SingleTaskOption {
	o.topic = topic
	return o
}
func (o *SingleTaskOption) WithMsg(msg ...[]byte) *SingleTaskOption {
	o.msg = msg
	return o
}
func (o *SingleTaskOption) WithRetry(retry int) *SingleTaskOption {
	o.retry = retry
	return o
}
func (o *SingleTaskOption) WithTimeout(timeout time.Duration) *SingleTaskOption {
	o.timeout = timeout
	return o
}
func (o *SingleTaskOption) WithCallback(cb func(error)) *SingleTaskOption {
	o.callback = cb
	return o
}
func (o *SingleTaskOption) WithDelay(d time.Duration) *SingleTaskOption {
	o.delay = d
	return o
}
