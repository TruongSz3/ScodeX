package events

import (
	"context"
	"errors"
	"sync"
	"time"
)

const DefaultSubscriberBuffer = 32

var ErrTopicRequired = errors.New("events: topic is required")

type Event struct {
	Topic     string
	Payload   any
	Timestamp time.Time
}

type PublishResult struct {
	Delivered int
	Dropped   int
}

type Bus struct {
	mu               sync.RWMutex
	nextSubscriberID uint64
	subscriberBuffer int
	subscribers      map[string]map[uint64]chan Event
}

func NewBus(subscriberBuffer int) *Bus {
	if subscriberBuffer <= 0 {
		subscriberBuffer = DefaultSubscriberBuffer
	}

	return &Bus{
		subscriberBuffer: subscriberBuffer,
		subscribers:      make(map[string]map[uint64]chan Event),
	}
}

func (b *Bus) Subscribe(ctx context.Context, topic string) (<-chan Event, func(), error) {
	if topic == "" {
		return nil, nil, ErrTopicRequired
	}

	ch := make(chan Event, b.subscriberBuffer)

	b.mu.Lock()
	b.nextSubscriberID++
	id := b.nextSubscriberID
	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[uint64]chan Event)
	}
	b.subscribers[topic][id] = ch
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		subs, ok := b.subscribers[topic]
		if !ok {
			b.mu.Unlock()
			return
		}
		subscriber, exists := subs[id]
		if !exists {
			b.mu.Unlock()
			return
		}
		delete(subs, id)
		if len(subs) == 0 {
			delete(b.subscribers, topic)
		}
		b.mu.Unlock()

		close(subscriber)
	}

	if ctx != nil {
		go func() {
			<-ctx.Done()
			unsubscribe()
		}()
	}

	return ch, unsubscribe, nil
}

func (b *Bus) Publish(ctx context.Context, topic string, payload any) (PublishResult, error) {
	if topic == "" {
		return PublishResult{}, ErrTopicRequired
	}
	if ctx != nil {
		select {
		case <-ctx.Done():
			return PublishResult{}, ctx.Err()
		default:
		}
	}

	event := Event{Topic: topic, Payload: payload, Timestamp: time.Now().UTC()}

	result := PublishResult{}
	b.mu.RLock()
	for _, subscriber := range b.subscribers[topic] {
		select {
		case subscriber <- event:
			result.Delivered++
		default:
			result.Dropped++
		}
	}
	b.mu.RUnlock()

	return result, nil
}
