package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"messanger/internal/entity"

	"github.com/segmentio/kafka-go"
)

const (
	saveRetries    = 3
	saveRetryDelay = time.Second
)

// Consumer читает сообщения из Kafka и сохраняет их в БД.
// Offset коммитится только после успешного сохранения (at-least-once).
type Consumer struct {
	reader   *kafka.Reader
	messages entity.MessageLogic
	log      *slog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, messages entity.MessageLogic, log *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10e6,
		}),
		messages: messages,
		log:      log,
	}
}

// Run обрабатывает сообщения, пока не отменён ctx.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		kafkaMsg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("fetch message from kafka: %w", err)
		}

		if err := c.handle(ctx, kafkaMsg); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		if err := c.reader.CommitMessages(ctx, kafkaMsg); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("commit offset: %w", err)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, kafkaMsg kafka.Message) error {
	var message entity.Message
	if err := json.Unmarshal(kafkaMsg.Value, &message); err != nil {
		// Битое сообщение не исправится при повторе — логируем и пропускаем.
		c.log.Error("skip invalid message", "offset", kafkaMsg.Offset, "error", err)
		return nil
	}

	var err error
	for attempt := 1; attempt <= saveRetries; attempt++ {
		err = c.messages.Save(ctx, message)
		if err == nil || errors.Is(err, entity.ErrInvalidInput) || attempt == saveRetries {
			break
		}

		c.log.Warn("save message failed, retrying", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(saveRetryDelay * time.Duration(attempt)):
		}
	}

	if errors.Is(err, entity.ErrInvalidInput) {
		c.log.Error("skip invalid message", "offset", kafkaMsg.Offset, "error", err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("save message (offset %d): %w", kafkaMsg.Offset, err)
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
