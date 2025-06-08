package repository

import (
	"context"
	"fmt"
	"messanger/internal/entity"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type messageRepository struct {
	poolPG *pgxpool.Pool
}

func NewMessageRepository(poolPG *pgxpool.Pool) entity.MessageRepository {
	return &messageRepository{
		poolPG: poolPG,
	}
}

func (m *messageRepository) Save(ctx context.Context, message entity.Message) error {
	sender, err := uuid.FromString(message.SenderID)
	if err != nil {
		return errors.Wrapf(err, "cant parsed id %v to uuid", message.SenderID)
	}
	receiver, err := uuid.FromString(message.ReceiverID)
	if err != nil {
		return errors.Wrapf(err, "cant parsed id %v to uuid", message.ReceiverID)
	}
	query := fmt.Sprint(`INSERT INTO messages (sender_id, receiver_id, "text", time_to_send) VALUES($1, $2, $3, $4)`)

	_, err = m.poolPG.Exec(ctx, query, sender, receiver, message.Text, message.TimeToSend)
	if err != nil {
		return errors.Wrapf(err, "cant save message from user %v to user %v", message.SenderID, message.ReceiverID)
	}
	return nil
}

func (m *messageRepository) AllByUserIDs(ctx context.Context, user1, user2 string) ([]entity.Message, error) {
	log := zap.NewExample().Sugar()
	userID1, err := uuid.FromString(user1)
	if err != nil {
		return nil, errors.Wrapf(err, "cant parse id %v to uuid", user1)
	}
	userID2, err := uuid.FromString(user2)
	if err != nil {
		return nil, errors.Wrapf(err, "cant parse id %v to uuid", user2)
	}

	query := fmt.Sprint(`SELECT m.sender_id, m.receiver_id, m."text", m.time_to_send FROM messages m
	WHERE (m.sender_id = $1 and m.receiver_id = $2) OR (m.sender_id = $2 and m.receiver_id = $1) 
	ORDER BY m.time_to_send DESC`)

	rows, err := m.poolPG.Query(ctx, query, userID1, userID2)
	if err != nil {
		return nil, errors.Wrapf(err, "cant get messages for %v and %v", user1, user2)
	}

	messages := make([]entity.Message, 0)
	for rows.Next() {
		var (
			msg        entity.Message
			senderID   uuid.UUID
			receiverID uuid.UUID
		)
		err = rows.Scan(&senderID, &receiverID, &msg.Text, &msg.TimeToSend)
		if err != nil {
			log.Error("cant scan message: %v", err)
			continue
		}
		msg.SenderID = senderID.String()
		msg.ReceiverID = receiverID.String()

		messages = append(messages, msg)
	}
	return messages, nil
}
