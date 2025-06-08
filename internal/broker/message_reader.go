package broker

import (
	"context"
	"encoding/json"
	"messanger/internal/entity"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type Reader struct {
	messageLogic entity.MessageLogic
	reader       *kafka.Reader
}

func (m *Reader) Shutdown(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewReader(ctx context.Context, address string, msgLogic entity.MessageLogic) *Reader {
	return &Reader{
		messageLogic: msgLogic,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     []string{address},
			Topic:       messageSendTopic,
			GroupID:     "my-group",
			StartOffset: kafka.LastOffset,
		})}
}

func (m *Reader) Start(ctx context.Context) error {
	for {
		msgKafka, err := m.reader.ReadMessage(ctx)
		if err != nil {
			return errors.Wrapf(err, "cant read message from topic %s", messageSendTopic)
		}

		if len(msgKafka.Value) > 0 {
			var message entity.Message
			err = json.Unmarshal(msgKafka.Value, &message)
			if err != nil {
				return errors.Wrapf(err, "cant unmarshal message from topic %s", messageSendTopic)
			}

			err = m.messageLogic.Save(ctx, message)
			if err != nil {
				return errors.Wrapf(err, "cant save message %v", message)
			}
		}
	}
}
