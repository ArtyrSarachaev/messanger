package entity

import "context"

type Message struct {
	UserID     int64  `json:"user_id"`
	Text       string `json:"text"`
	TimeToSend int64  `json:"time_to_send"`
}

type SendMessageBroker interface {
	SendMessage(ctx context.Context, message Message) error
}

type Writer interface {
	Write(ctx context.Context, topic string, data interface{}) error
}
