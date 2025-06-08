package logic

import (
	"context"
	"messanger/internal/entity"

	"github.com/pkg/errors"
)

const (
	formatTimeForResponse = "2006-01-02 15:04:05"
)

type messageLogic struct {
	userLogic   entity.UserLogic
	messageRepo entity.MessageRepository
}

func NewMessageLogic(msgRepo entity.MessageRepository, userLogic entity.UserLogic) entity.MessageLogic {
	return &messageLogic{
		userLogic:   userLogic,
		messageRepo: msgRepo,
	}
}

func (m *messageLogic) ByName(ctx context.Context, username string) ([]entity.MessageForResponse, error) {
	user1, err := m.userLogic.ByFullName(ctx, username)
	if err != nil || user1.ID == "" {
		return nil, errors.Wrapf(err, "cant get user by username %v", username)
	}

	userID2, ok := ctx.Value(entity.UserIDKey).(string)
	if !ok {
		return nil, errors.New("cant get user id from context")
	}

	user2, err := m.userLogic.ByUserID(ctx, userID2)
	if err != nil || user2.ID == "" || user2.Username == username {
		return nil, errors.Wrapf(err, "cant get user by user id %v", userID2)
	}

	messages, err := m.messageRepo.AllByUserIDs(ctx, user1.ID, user2.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "cant get all messages by users %v", username)
	}

	messagesResp := make([]entity.MessageForResponse, 0, len(messages))
	for _, messageFromDB := range messages {
		messageResp := entity.MessageForResponse{
			Text:       messageFromDB.Text,
			TimeToSend: messageFromDB.TimeToSend.Format(formatTimeForResponse),
		}

		if messageFromDB.SenderID == user1.ID {
			messageResp.SenderName = user1.Username
			messageResp.ReceiverName = user2.Username
		} else {
			messageResp.SenderName = user2.Username
			messageResp.ReceiverName = user1.Username
		}

		messagesResp = append(messagesResp, messageResp)
	}

	return messagesResp, nil
}

func (m *messageLogic) Save(ctx context.Context, message entity.Message) error {
	return m.messageRepo.Save(ctx, message)
}
