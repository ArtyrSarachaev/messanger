package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"messanger/internal/entity"
	"messanger/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type messageReader struct {
	r *kafka.Reader
}

func StartMessageKafkaReader(ctx context.Context, address string) error {
	log := logger.LoggerFromContext(ctx)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{address},
		Topic:   messageSendTopic,
	})

	for {
		msgB, err := reader.ReadMessage(ctx)
		if err != nil {
			log.Errorf("cant read message from topic %s, with error %v", messageSendTopic, err)
			break
		}
		var message entity.Message
		json.Unmarshal(msgB.Value, &message)
		fmt.Println(message)
	}

	return reader.Close()
}
