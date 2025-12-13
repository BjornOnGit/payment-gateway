package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQBus struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	url  string
}

// NewRabbitMQBus creates a new RabbitMQ bus instance
func NewRabbitMQBus(url string) (*RabbitMQBus, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return &RabbitMQBus{
		conn: conn,
		ch:   ch,
		url:  url,
	}, nil
}

// Publish publishes a message to a topic
func (b *RabbitMQBus) Publish(ctx context.Context, topic, key string, payload []byte) error {
	// Declare the exchange (topic-based)
	err := b.ch.ExchangeDeclare(
		topic,   // name
		"topic", // kind
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Publish the message with routing key
	// For topic exchange, use the key as routing key if provided, otherwise use topic
	routingKey := key
	if routingKey == "" {
		routingKey = topic
	}

	err = b.ch.PublishWithContext(
		ctx,
		topic,      // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Subscribe subscribes to a topic with a handler function
func (b *RabbitMQBus) Subscribe(ctx context.Context, topic string, handler bus.HandlerFn) error {
	// Declare the exchange
	err := b.ch.ExchangeDeclare(
		topic,   // name
		"topic", // kind
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare DLQ exchange for dead letters
	dlqExchange := "dlq." + topic
	err = b.ch.ExchangeDeclare(
		dlqExchange, // name
		"topic",     // kind
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ exchange: %w", err)
	}

	// Declare DLQ queue
	dlqQueue, err := b.ch.QueueDeclare(
		dlqExchange, // name - use exchange name for DLQ queue
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare DLQ queue: %w", err)
	}

	// Bind DLQ queue to DLQ exchange
	err = b.ch.QueueBind(
		dlqQueue.Name, // queue name
		"#",           // routing key
		dlqExchange,   // exchange
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind DLQ queue: %w", err)
	}

	// Declare a queue unique to this consumer with DLQ configuration
	q, err := b.ch.QueueDeclare(
		"",    // name - empty string means auto-generate a unique queue name
		false, // durable
		true,  // delete when unused
		true,  // exclusive - only this connection can use it
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": dlqExchange, // Send rejected messages to DLQ
			"x-message-ttl":          300000,      // 5 minutes TTL for retry messages
		}, // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind the queue to the exchange using wildcard binding (match all routing keys)
	err = b.ch.QueueBind(
		q.Name, // queue name
		"#",    // routing key - # matches all routing keys in topic exchange
		topic,  // exchange
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming messages
	msgs, err := b.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	// Start goroutine to handle messages
	go func() {
		for m := range msgs {
			// Create a new context for each message (not tied to HTTP context)
			msgCtx := context.Background()

			// Track retry count from x-death header
			retryCount := 0
			if m.Headers != nil {
				if xDeath, ok := m.Headers["x-death"].([]interface{}); ok && len(xDeath) > 0 {
					if death, ok := xDeath[0].(amqp.Table); ok {
						if count, ok := death["count"].(int64); ok {
							retryCount = int(count)
						}
					}
				}
			}

			// Store retry count in context for handler to use
			msgCtx = context.WithValue(msgCtx, "retry_count", retryCount)

			// RabbitMQ routing key is used as the key parameter
			if err := handler(msgCtx, topic, m.RoutingKey, m.Body); err != nil {
				// Check if we've exceeded max retries (3 attempts)
				if retryCount >= 3 {
					// Send to DLQ by rejecting without requeue
					log.Printf("[%s] max retries exceeded (%d), sending to DLQ: %v", topic, retryCount, err)
					m.Nack(false, false) // don't requeue, will go to DLQ
				} else {
					// Requeue for retry
					log.Printf("[%s] handler error (retry %d/3): %v", topic, retryCount, err)
					m.Nack(false, true) // requeue for retry
				}
				continue
			}

			// ACK the message on success
			m.Ack(false)
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (b *RabbitMQBus) Close() error {
	if b.ch != nil {
		b.ch.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}
