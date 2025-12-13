package bus

import "context"

type HandlerFn func(ctx context.Context, topic string, key string, payload []byte) error

type Producer interface {
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}

type Consumer interface {
	Subscribe(ctx context.Context, topic string, handler HandlerFn) error
	Close() error
}

type Bus interface {
	Producer
	Consumer
}