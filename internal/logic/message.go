package logic

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"messanger/internal/entity"
)

const (
	DefaultDialogLimit = 50
	MaxDialogLimit     = 200
)

type messageLogic struct {
	repo      entity.MessageRepository
	userLogic entity.UserLogic
}

func NewMessageLogic(repo entity.MessageRepository, userLogic entity.UserLogic) entity.MessageLogic {
	return &messageLogic{repo: repo, userLogic: userLogic}
}

// Save сохраняет сообщение, прочитанное из Kafka.
func (m *messageLogic) Save(ctx context.Context, message entity.Message) error {
	if err := ValidateMessage(message); err != nil {
		return err
	}
	return m.repo.Save(ctx, message)
}

// Dialog возвращает переписку текущего пользователя с companion (по username),
// от новых сообщений к старым.
func (m *messageLogic) Dialog(ctx context.Context, currentUserID, companion string, limit, offset int) ([]entity.MessageView, error) {
	limit, offset = normalizePage(limit, offset)

	me, err := m.userLogic.ByID(ctx, currentUserID)
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}

	other, err := m.userLogic.ByUsername(ctx, companion)
	if err != nil {
		return nil, err
	}

	messages, err := m.repo.Dialog(ctx, me.ID, other.ID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get dialog: %w", err)
	}

	names := map[string]string{me.ID: me.Username, other.ID: other.Username}

	views := make([]entity.MessageView, 0, len(messages))
	for _, msg := range messages {
		views = append(views, entity.MessageView{
			From:      names[msg.SenderID],
			To:        names[msg.ReceiverID],
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt,
		})
	}

	return views, nil
}

func ValidateMessage(message entity.Message) error {
	if message.SenderID == "" || message.ReceiverID == "" {
		return fmt.Errorf("%w: sender and receiver are required", entity.ErrInvalidInput)
	}
	if strings.TrimSpace(message.Text) == "" {
		return fmt.Errorf("%w: message is empty", entity.ErrInvalidInput)
	}
	if utf8.RuneCountInString(message.Text) > entity.MaxMessageLength {
		return fmt.Errorf("%w: message is longer than %d characters", entity.ErrInvalidInput, entity.MaxMessageLength)
	}
	return nil
}

func normalizePage(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = DefaultDialogLimit
	}
	if limit > MaxDialogLimit {
		limit = MaxDialogLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
