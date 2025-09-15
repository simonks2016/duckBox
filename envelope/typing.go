package envelope

import "fmt"

type EventType int

const (
	CommandEvent EventType = iota
	DomainEvent
	IntegrationEvent
)

func (e EventType) String() string {
	switch e {
	case CommandEvent:
		return "cmd"
	case DomainEvent:
		return "d.evt"
	case IntegrationEvent:
		return "i.evt"
	default:
		return "other.evt"
	}
}

func MakeEventKind(et EventType, boundedContext, eventName string) string {
	return fmt.Sprintf("%s.%s.%s", et.String(), boundedContext, eventName)
}
