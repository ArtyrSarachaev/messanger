package http

import (
	"messanger/internal/entity"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MessagesHandler struct {
	messageLogic entity.MessageLogic
}

func NewMessageHandler(rg *echo.Group, messageLogic entity.MessageLogic) {
	m := MessagesHandler{
		messageLogic: messageLogic,
	}
	rg.GET("/", m.getMessages)
}

func (m *MessagesHandler) getMessages(c echo.Context) error {
	ctx := c.Request().Context()

	username := c.QueryParam("username")
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	messages, err := m.messageLogic.ByName(ctx, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	return c.JSON(http.StatusOK, messages)
}
