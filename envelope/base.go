package envelope

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type Envelope[payload any] struct {
	EventId      string  `json:"event_id"`
	EventKind    string  `json:"event_kind"`
	EventVersion string  `json:"event_version"`
	OccurredAt   int64   `json:"occurred_at"`
	Payload      payload `json:"payload"`
}

func (e *Envelope[payload]) Decode(data []byte) (payload, error) {
	var nilPayload payload
	// JSON 转码
	if err := json.Unmarshal(data, e); err != nil {
		return nilPayload, err
	}
	return e.Payload, nil
}

func (e *Envelope[payload]) Encode() ([]byte, error) {
	marshal, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}

func (e *Envelope[payload]) SetVersion(version string) *Envelope[payload] {
	e.EventVersion = version
	return e
}

func NewEnvelope[T any](EventKind string, payload T) *Envelope[T] {

	return &Envelope[T]{
		EventKind:    EventKind,
		EventId:      uuid.NewString(),
		EventVersion: "1.0",
		OccurredAt:   time.Now().Unix(),
		Payload:      payload,
	}
}

func LoadEnvelope[T any](data []byte) (*Envelope[T], error) {
	var e Envelope[T]
	// 转码JSON
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}
