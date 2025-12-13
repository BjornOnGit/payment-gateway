package bus

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type memoryBus struct {
	subs map[string][]HandlerFn
	lock sync.RWMutex
	closed bool
}

func NewInMemoryBus() Bus {
	return &memoryBus{
		subs: make(map[string][]HandlerFn),
	}
}

func (m *memoryBus) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	m.lock.RLock()
	handlers := append([]HandlerFn(nil), m.subs[topic]...)
	m.lock.RUnlock()

	if len(handlers)  == 0{
		return nil
	}

	for _, h := range handlers {
		h := h
		go func ()  {
			ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			if err := h(ctx2, topic, key, payload); err != nil {
				log.Printf("[bus] handler error for topic=%s: %v", topic, err)
			}
		}()
	}
	return nil
}

func (m *memoryBus) Subscribe(ctx context.Context, topic string, handler HandlerFn) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.closed {
		return errors.New("bus closed")
	}
	m.subs[topic] = append(m.subs[topic], handler)
	return nil
}

func (m *memoryBus) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.closed = true
	m.subs = make(map[string][]HandlerFn)
	return nil
}