package broker

import (
	"context"
	"encoding/json"
	"messanger/internal/entity"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type messageWriter struct {
	w *kafka.Writer
}

func NewWriter(address string) entity.MessageKafkaBroker {
	return &messageWriter{
		w: &kafka.Writer{
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			Addr:                   kafka.TCP(address),
		},
	}
}

func (k *messageWriter) Send(ctx context.Context, message entity.Message) error {
	senderID, ok := ctx.Value(entity.UserIDKey).(string)
	if !ok {
		return errors.New("cant get username from context")
	}
	message.SenderID = senderID
	messageToKafka, err := json.Marshal(message)
	if err != nil {
		return errors.Wrapf(err, "cant marshal message %v", message)
	}
	k.w.Topic = messageSendTopic
	return k.w.WriteMessages(ctx,
		kafka.Message{
			Key:   []byte(senderID),
			Value: messageToKafka,
		})
}
