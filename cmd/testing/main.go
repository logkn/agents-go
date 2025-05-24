package main

import "time"

type Model any

type Agent struct {
	model        Model
	name         string
	instructions string
}

type Message any

type Input struct {
	OfString   string
	OfMessages []Message
}

type EventBus struct {
	Events  chan Event
	proxies map[EventType]chan Event
}

func NewEventBus() EventBus {
	return EventBus{
		Events:  make(chan Event),
		proxies: make(map[EventType]chan Event),
	}
}

func (bus *EventBus) ensureProxy(eventType EventType) {
	if _, ok := bus.proxies[eventType]; !ok {
		bus.proxies[eventType] = make(chan Event)
	}
}

func (bus *EventBus) Send(event Event) {
	// send to central bus
	bus.Events <- event

	// send to proxy channel
	eventType := event.EventType()
	bus.ensureProxy(eventType)

	bus.proxies[eventType] <- event
}

func (bus *EventBus) Subscribe(eventType EventType) <-chan Event {
	// ensure the proxy channel exists
	bus.ensureProxy(eventType)
	// return the proxy channel
	return bus.proxies[eventType]
}

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

type Response any

type RunResult interface {
	Events() chan<- Event
	Response() Response
	NextInput() Input
}

type Runner[State any] interface {
	Run(agent Agent, input Input, state *State) RunResult
}
