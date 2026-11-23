package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"messanger/internal/entity"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{}, // сообщения одного диалога попадают в одну партицию
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			BatchTimeout:           10 * time.Millisecond,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, message entity.Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(dialogKey(message.SenderID, message.ReceiverID)),
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("write message to kafka: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

// dialogKey одинаков для обоих направлений диалога — порядок сообщений
// внутри диалога сохраняется.
func dialogKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + ":" + b
}
