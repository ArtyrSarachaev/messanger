package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"messanger/internal/entity"
	"messanger/internal/logic"
	"messanger/internal/middleware"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	pongWait     = 60 * time.Second
	pingInterval = pongWait * 9 / 10
	maxFrameSize = 16 * 1024
)

// Исходящие кадры для клиента.
const (
	frameMessage = "message" // входящее сообщение от собеседника
	frameSent    = "sent"    // подтверждение, что сообщение принято
	frameError   = "error"
)

var errPublishFailed = errors.New("message was not sent, try again later")

type outgoingFrame struct {
	Type    string              `json:"type"`
	Message *entity.MessageView `json:"message,omitempty"`
	Error   string              `json:"error,omitempty"`
}

type Server struct {
	echo      *echo.Echo
	hub       *hub
	upgrader  websocket.Upgrader
	users     entity.UserLogic
	publisher entity.MessagePublisher
	log       *slog.Logger
}

func NewServer(users entity.UserLogic, publisher entity.MessagePublisher, tokens entity.TokenManager, log *slog.Logger) *Server {
	s := &Server{
		echo:      echo.New(),
		hub:       newHub(),
		upgrader:  websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024},
		users:     users,
		publisher: publisher,
		log:       log,
	}

	s.echo.HideBanner = true
	s.echo.HidePort = true
	s.echo.GET("/ws", s.connect, middleware.JWT(tokens))

	return s
}

func (s *Server) Start(address string) error {
	return s.echo.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Hijacked-соединения не закрываются при Shutdown http-сервера, закрываем их сами.
	s.hub.closeAll()
	return s.echo.Shutdown(ctx)
}

func (s *Server) connect(c echo.Context) error {
	userID, ok := entity.UserIDFromContext(c.Request().Context())
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	sender, err := s.users.ByID(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unknown user")
	}

	conn, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		// Upgrade уже отправил клиенту ответ с ошибкой.
		s.log.Warn("websocket upgrade failed", "error", err)
		return nil
	}

	cl := &client{userID: sender.ID, conn: conn}
	s.hub.add(cl)
	s.log.Info("websocket connected", "user", sender.Username, "remote", conn.RemoteAddr().String())

	done := make(chan struct{})
	go s.keepAlive(cl, done)

	defer func() {
		close(done)
		s.hub.remove(cl)
		_ = conn.Close()
		s.log.Info("websocket disconnected", "user", sender.Username)
	}()

	s.readLoop(cl, sender)

	return nil
}

// Контекст запроса после Upgrade не используем: для долгоживущего соединения
// каждое сообщение обрабатывается со своим таймаутом.
func (s *Server) readLoop(cl *client, sender entity.User) {
	cl.conn.SetReadLimit(maxFrameSize)
	_ = cl.conn.SetReadDeadline(time.Now().Add(pongWait))
	cl.conn.SetPongHandler(func(string) error {
		return cl.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var in entity.IncomingMessage
		messageType, data, err := cl.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				s.log.Warn("websocket read failed", "user", sender.Username, "error", err)
			}
			return
		}
		if messageType != websocket.TextMessage {
			s.sendError(cl, fmt.Errorf("%w: only text frames are supported", entity.ErrInvalidInput))
			continue
		}
		// Невалидный JSON не рвёт соединение — клиент получает ошибку и может продолжать.
		if err := json.Unmarshal(data, &in); err != nil {
			s.sendError(cl, fmt.Errorf("%w: invalid JSON, expected {\"to\": \"username\", \"text\": \"...\"}", entity.ErrInvalidInput))
			continue
		}

		s.handleMessage(cl, sender, in)
	}
}

func (s *Server) handleMessage(cl *client, sender entity.User, in entity.IncomingMessage) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	receiver, err := s.users.ByUsername(ctx, in.To)
	if err != nil {
		s.sendError(cl, err)
		return
	}

	msg := entity.Message{
		SenderID:   sender.ID,
		ReceiverID: receiver.ID,
		Text:       in.Text,
		CreatedAt:  time.Now().UTC(),
	}
	if err := logic.ValidateMessage(msg); err != nil {
		s.sendError(cl, err)
		return
	}

	// Сообщение сохраняется асинхронно: пишем в Kafka, консьюмер кладёт в Postgres.
	if err := s.publisher.Publish(ctx, msg); err != nil {
		s.log.Error("publish message failed", "error", err)
		s.sendError(cl, errPublishFailed)
		return
	}

	view := entity.MessageView{
		From:      sender.Username,
		To:        receiver.Username,
		Text:      msg.Text,
		CreatedAt: msg.CreatedAt,
	}

	// Доставляем получателю во все его открытые соединения.
	for _, rc := range s.hub.connections(receiver.ID) {
		if err := rc.writeJSON(outgoingFrame{Type: frameMessage, Message: &view}); err != nil {
			s.log.Warn("deliver message failed", "user", receiver.Username, "error", err)
		}
	}

	// Отправителю — подтверждение в текущее соединение и копию в остальные его вкладки.
	for _, sc := range s.hub.connections(sender.ID) {
		frameType := frameMessage
		if sc == cl {
			frameType = frameSent
		}
		if sender.ID == receiver.ID && sc != cl {
			continue // сообщение самому себе уже доставлено выше
		}
		if err := sc.writeJSON(outgoingFrame{Type: frameType, Message: &view}); err != nil {
			s.log.Warn("write to sender failed", "user", sender.Username, "error", err)
		}
	}
}

func (s *Server) sendError(cl *client, err error) {
	text := "internal error"
	switch {
	case errors.Is(err, entity.ErrUserNotFound):
		text = "receiver not found"
	case errors.Is(err, entity.ErrInvalidInput), errors.Is(err, errPublishFailed):
		text = err.Error()
	default:
		s.log.Error("handle websocket message failed", "error", err)
	}

	if writeErr := cl.writeJSON(outgoingFrame{Type: frameError, Error: text}); writeErr != nil {
		s.log.Warn("write error frame failed", "error", writeErr)
	}
}

func (s *Server) keepAlive(cl *client, done <-chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := cl.writeControl(websocket.PingMessage); err != nil {
				_ = cl.conn.Close()
				return
			}
		}
	}
}
