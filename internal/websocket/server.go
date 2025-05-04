package websocket

import (
	"messanger/internal/entity"
	"messanger/internal/middleware"

	"github.com/labstack/echo/v4"
)

func StartWSServer(sendKafka entity.SendMessageBroker) *echo.Echo {
	e := echo.New()

	e.Use(
		middleware.JWTMiddleware,
		middleware.AddLoggerInHandlerContext,
	)
	NewStartWSHandlers(e.Group("/send"), sendKafka)

	return e
}
