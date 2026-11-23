package repository

import (
	"context"
	"fmt"

	"messanger/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type messageRepository struct {
	pool *pgxpool.Pool
}

func NewMessageRepository(pool *pgxpool.Pool) entity.MessageRepository {
	return &messageRepository{pool: pool}
}

func (r *messageRepository) Save(ctx context.Context, message entity.Message) error {
	const query = `
		INSERT INTO messages (sender_id, receiver_id, text, created_at)
		VALUES ($1::uuid, $2::uuid, $3, $4)`

	_, err := r.pool.Exec(ctx, query, message.SenderID, message.ReceiverID, message.Text, message.CreatedAt)
	if err != nil {
		return fmt.Errorf("save message from %s to %s: %w", message.SenderID, message.ReceiverID, err)
	}

	return nil
}

func (r *messageRepository) Dialog(ctx context.Context, userID1, userID2 string, limit, offset int) ([]entity.Message, error) {
	const query = `
		SELECT sender_id::text, receiver_id::text, text, created_at
		FROM messages
		WHERE (sender_id = $1::uuid AND receiver_id = $2::uuid)
		   OR (sender_id = $2::uuid AND receiver_id = $1::uuid)
		ORDER BY created_at DESC, id DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, userID1, userID2, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query dialog %s<->%s: %w", userID1, userID2, err)
	}

	messages, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.Message, error) {
		var msg entity.Message
		err := row.Scan(&msg.SenderID, &msg.ReceiverID, &msg.Text, &msg.CreatedAt)
		return msg, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan dialog %s<->%s: %w", userID1, userID2, err)
	}

	return messages, nil
}
