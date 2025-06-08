package entity

import (
	"context"
	"time"
)

type MessageWebSocket struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

type Message struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	TimeToSend time.Time `json:"time_to_send"`
}

type MessageForResponse struct {
	SenderName   string `json:"sender_id"`
	ReceiverName string `json:"receiver_id"`
	Text         string `json:"text"`
	TimeToSend   string `json:"time_to_send"`
}

type MessageRepository interface {
	Save(ctx context.Context, message Message) error
	AllByUserIDs(ctx context.Context, userID1, userID2 string) ([]Message, error)
}

type MessageLogic interface {
	Save(ctx context.Context, message Message) error
	ByName(ctx context.Context, username string) ([]MessageForResponse, error)
}

type MessageKafkaBroker interface {
	Send(ctx context.Context, message Message) error
}

type Writer interface {
	Write(ctx context.Context, topic string, data interface{}) error
}
