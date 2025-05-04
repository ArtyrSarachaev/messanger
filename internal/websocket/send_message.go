package websocket

import (
	"context"
	"messanger/internal/entity"
	env "messanger/pkg/environment"
	"messanger/pkg/logger"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type wsSendMessage struct {
	wsUpg *websocket.Upgrader
	connections
	sendMessageKafka entity.SendMessageBroker
}

type connections struct {
	mutex     *sync.RWMutex
	wsClients map[int64]*websocket.Conn
}

func NewStartWSHandlers(rg *echo.Group, sendKafka entity.SendMessageBroker) {
	w := wsSendMessage{
		wsUpg: &websocket.Upgrader{},
		connections: connections{
			mutex:     &sync.RWMutex{},
			wsClients: make(map[int64]*websocket.Conn),
		},
		sendMessageKafka: sendKafka,
	}

	rg.GET("/", w.send)
}

func (w *wsSendMessage) send(c echo.Context) error {
	ctx := c.Request().Context()
	log := logger.LoggerFromContext(ctx)
	conn, err := w.wsUpg.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID := env.GetUserId(ctx)
	defer func(userID int64) {
		w.connections.mutex.Lock()
		delete(w.connections.wsClients, userID)
		w.connections.mutex.Unlock()

		conn.Close()
	}(userID)

	log.Infof("upgrade connection: %v, for user with id: %v", conn.RemoteAddr().String(), userID)

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
	log := logger.LoggerFromContext(ctx)
	var msg entity.Message
	err := conn.ReadJSON(&msg)
	if err != nil {
		wsError, ok := err.(*websocket.CloseError)
		if !ok || wsError.Code != websocket.CloseGoingAway {
			log.Errorf("error while reading from websocket: %v", err)
		}
	}
	msg.TimeToSend = time.Now().Unix()
	return msg, nil
}

func (w *wsSendMessage) writeToClient(ctx context.Context, conn *websocket.Conn, msg entity.Message) error {
	log := logger.LoggerFromContext(ctx)
	err := w.sendMessageKafka.SendMessage(ctx, msg)
	if err != nil {
		log.Errorf("error to write message in kafka: %v", err)
	}

	w.connections.mutex.RLock()
	defer w.connections.mutex.RUnlock()
	conn, ok := w.connections.wsClients[msg.UserID]
	if ok {
		if err := conn.WriteJSON(msg.Text); err != nil {
			log.Errorf("error with writing message: %v", err)
		}
	}
	return nil
}
