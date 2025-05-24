package events

import "time"

type EventType string

type EventVariant interface {
	EventType() EventType
}

type Event struct {
	Timestamp time.Time
	Payload   EventVariant
}

func (e Event) EventType() EventType {
	return e.Payload.EventType()
}

func NewEvent(payload EventVariant) Event {
	return Event{
		Timestamp: time.Now(),
		Payload:   payload,
	}
}
