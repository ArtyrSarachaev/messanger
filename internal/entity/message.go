package entity

import (
	"context"
	"time"
)

const MaxMessageLength = 4096

// IncomingMessage — сообщение, которое клиент отправляет в WebSocket.
type IncomingMessage struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

// Message — сообщение во внутреннем представлении (Kafka, БД).
type Message struct {
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	CreatedAt  time.Time `json:"created_at"`
}

// MessageView — сообщение в ответах API и в WebSocket.
type MessageView struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageRepository interface {
	Save(ctx context.Context, message Message) error
	Dialog(ctx context.Context, userID1, userID2 string, limit, offset int) ([]Message, error)
}

type MessageLogic interface {
	Save(ctx context.Context, message Message) error
	Dialog(ctx context.Context, currentUserID, companion string, limit, offset int) ([]MessageView, error)
}

type MessagePublisher interface {
	Publish(ctx context.Context, message Message) error
}
