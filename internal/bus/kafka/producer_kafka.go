package kafka

import (
    "context"
    "github.com/IBM/sarama"
)

type KafkaProducer struct {
    producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) (*KafkaProducer, error) {
    cfg := sarama.NewConfig()
    cfg.Producer.Return.Successes = true
    cfg.Producer.Return.Errors = true
    cfg.Producer.RequiredAcks = sarama.WaitForAll
    cfg.Producer.Retry.Max = 5

    p, err := sarama.NewSyncProducer(brokers, cfg)
    if err != nil {
        return nil, err
    }

    return &KafkaProducer{producer: p}, nil
}

func (kp *KafkaProducer) Publish(ctx context.Context, topic, key string, payload []byte) error {
    msg := &sarama.ProducerMessage{
        Topic: topic,
        Key:   sarama.StringEncoder(key),
        Value: sarama.ByteEncoder(payload),
    }
    _, _, err := kp.producer.SendMessage(msg)
    return err
}
