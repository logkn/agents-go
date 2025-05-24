package events

type EventBus struct {
	events  chan Event
	proxies map[EventType]chan Event
}

func NewEventBus() EventBus {
	return EventBus{
		events:  make(chan Event),
		proxies: make(map[EventType]chan Event),
	}
}

func (bus *EventBus) ensureProxy(eventType EventType) {
	if _, ok := bus.proxies[eventType]; !ok {
		bus.proxies[eventType] = make(chan Event)
	}
}

func (bus *EventBus) SendEvent(event Event) {
	// send to central bus
	bus.events <- event

	// send to proxy channel
	eventType := event.EventType()
	bus.ensureProxy(eventType)

	bus.proxies[eventType] <- event
}

func (bus *EventBus) SendVariant(eventVar EventVariant) {
	// make from variant
	event := NewEvent(eventVar)
	bus.SendEvent(event)
}

func (bus *EventBus) ListenToType(eventType EventType) <-chan Event {
	// ensure the proxy channel exists
	bus.ensureProxy(eventType)
	// return the proxy channel
	return bus.proxies[eventType]
}

func (bus *EventBus) ListenAll() <-chan Event {
	// return the central bus channel
	return bus.events
}
