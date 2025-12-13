package rabbitmq

import (
	"context"
	"log"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitConsumer struct {
	ch *amqp.Channel
}

func NewRabbitConsumer(url string) (*RabbitConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RabbitConsumer{ch: ch}, nil
}

func (c *RabbitConsumer) Subscribe(ctx context.Context, queue string, handler bus.HandlerFn) error {
	msgs, err := c.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for m := range msgs {
			ctx := context.Background()

			// RabbitMQ message routing key is used as the key parameter
			// Queue name is used as the topic parameter
			if err := handler(ctx, queue, m.RoutingKey, m.Body); err != nil {
				// don't ACK on error — worker can retry or dead-letter later
				log.Println("Handler error:", err)
				continue
			}

			m.Ack(false)
		}
	}()
	return nil
}
