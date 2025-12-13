package rabbitmq

import (
    "context"

    amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitProducer struct {
    ch *amqp.Channel
}

func NewRabbitProducer(url string) (*RabbitProducer, error) {
    conn, err := amqp.Dial(url)
    if err != nil {
        return nil, err
    }
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    return &RabbitProducer{ch: ch}, nil
}

func (p *RabbitProducer) Publish(ctx context.Context, topic, key string, payload []byte) error {
    return p.ch.PublishWithContext(ctx,
        topic, // Exchange
        key,   // Routing key
        false, false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        payload,
        },
    )
}
