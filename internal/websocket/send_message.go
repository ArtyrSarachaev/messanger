package websocket

import (
	"context"
	"messanger/internal/entity"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type wsSendMessage struct {
	wsUpg *websocket.Upgrader
	connections
	messageKafka entity.MessageKafkaBroker
	userLogic    entity.UserLogic
}

type connections struct {
	mutex     *sync.RWMutex
	wsClients map[string]*websocket.Conn
}

func NewStartWSHandlers(rg *echo.Group, sendKafka entity.MessageKafkaBroker, userLogic entity.UserLogic) {
	w := wsSendMessage{
		wsUpg: &websocket.Upgrader{},
		connections: connections{
			mutex:     &sync.RWMutex{},
			wsClients: make(map[string]*websocket.Conn),
		},
		messageKafka: sendKafka,
		userLogic:    userLogic,
	}

	rg.GET("/", w.connect)
}

func (w *wsSendMessage) connect(c echo.Context) error {
	ctx := c.Request().Context()
	log := zap.NewExample().Sugar()
	conn, err := w.wsUpg.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID, ok := ctx.Value(entity.UserIDKey).(string)
	if !ok {
		return errors.New("cant get username from context")
	}
	defer func(userID string) {
		w.connections.mutex.Lock()
		delete(w.connections.wsClients, userID)
		w.connections.mutex.Unlock()

		conn.Close()
	}(userID)

	log.Infof("upgrade connection: %v, for user %v", conn.RemoteAddr().String(), userID)

	w.connections.mutex.Lock()
	w.connections.wsClients[userID] = conn
	w.connections.mutex.Unlock()

	for {
		msg, err := w.readFromClient(ctx, w.wsClients[userID])
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = w.writeToClient(ctx, w.wsClients[userID], msg)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}
}

func (w *wsSendMessage) readFromClient(ctx context.Context, conn *websocket.Conn) (entity.Message, error) {
	log := zap.NewExample().Sugar()
	var messageFromWS entity.MessageWebSocket
	err := conn.ReadJSON(&messageFromWS)
	if err != nil {
		wsError, ok := err.(*websocket.CloseError)
		if !ok || wsError.Code != websocket.CloseGoingAway {
			log.Errorf("error while reading from websocket: %v", err)
		}
	}

	receiver, err := w.userLogic.ByFullName(ctx, messageFromWS.Username)
	if err != nil {
		return entity.Message{}, errors.Wrapf(err, "cant get user by name %v", messageFromWS.Username)
	}

	senderID, ok := ctx.Value(entity.UserIDKey).(string)
	if !ok {
		return entity.Message{}, errors.Wrap(err, "cant get sender id from context")
	}

	return entity.Message{
		SenderID:   senderID,
		ReceiverID: receiver.ID,
		Text:       messageFromWS.Text,
		TimeToSend: time.Now(),
	}, nil
}

func (w *wsSendMessage) writeToClient(ctx context.Context, conn *websocket.Conn, message entity.Message) error {
	log := zap.NewExample().Sugar()
	if message.ReceiverID == "" {
		return errors.New("unknown receiver message")
	}
	if message.Text == "" {
		return errors.New("message is empty")
	}
	err := w.messageKafka.Send(ctx, message)
	if err != nil {
		log.Errorf("error to write message in kafka: %v", err)
	}

	w.connections.mutex.RLock()
	defer w.connections.mutex.RUnlock()
	conn, ok := w.connections.wsClients[message.ReceiverID]
	if ok {
		if err := conn.WriteJSON(message.Text); err != nil {
			log.Errorf("error with writing message: %v", err)
		}
	}
	return nil
}
