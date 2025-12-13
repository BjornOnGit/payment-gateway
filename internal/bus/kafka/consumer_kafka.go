package kafka

import (
	"context"
	"log"

	"github.com/BjornOnGit/payment-gateway/internal/bus"
	"github.com/IBM/sarama"
)

type KafkaConsumer struct {
	group sarama.ConsumerGroup
}

func NewKafkaConsumer(brokers []string, groupID string) (*KafkaConsumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V3_6_0_0
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, err
	}
	return &KafkaConsumer{group: group}, nil
}

func (kc *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler bus.HandlerFn) error {
	go func() {
		for {
			err := kc.group.Consume(context.Background(), []string{topic}, &consumerHandler{handler})
			if err != nil {
				log.Println("Kafka consume error:", err)
			}
		}
	}()
	return nil
}

type consumerHandler struct {
	handler bus.HandlerFn
}

func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		ctx := context.Background()

		// Extract key from Kafka message (convert []byte to string)
		key := string(msg.Key)
		topic := msg.Topic

		if err := h.handler(ctx, topic, key, msg.Value); err != nil {
			// No auto-retry — leave record uncommitted
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
