package broker

import (
	"context"
	"encoding/json"
	"messanger/internal/entity"
	"messanger/pkg/logger"

	env "messanger/pkg/environment"

	"github.com/segmentio/kafka-go"
)

type messageWriter struct {
	w *kafka.Writer
}

func NewMessageKafkaWriter(address string) entity.SendMessageBroker {
	return &messageWriter{
		w: &kafka.Writer{
			RequiredAcks:           kafka.RequireAll,
			AllowAutoTopicCreation: true,
			Addr:                   kafka.TCP(address),
		},
	}
}

func (k *messageWriter) SendMessage(ctx context.Context, message entity.Message) error {
	log := logger.LoggerFromContext(ctx)
	//todo надо сделать человеческое реквест айди
	userID, err := json.Marshal(env.GetUserId(ctx))
	if err != nil {
		log.Errorf("cant marshal key %v, with error %v", env.GetUserId(ctx), err)
		return err
	}
	messageToKafka, err := json.Marshal(message)
	if err != nil {
		log.Errorf("cant marshal message %v, with error %v", message, err)
		return err
	}
	k.w.Topic = messageSendTopic
	return k.w.WriteMessages(ctx,
		kafka.Message{
			Key:   userID,
			Value: messageToKafka,
		})
}
