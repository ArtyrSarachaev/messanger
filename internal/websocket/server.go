package websocket

import (
	"messanger/internal/entity"
	"messanger/internal/middleware"

	"github.com/labstack/echo/v4"
)

func Server(sendKafka entity.MessageKafkaBroker, userLogic entity.UserLogic) *echo.Echo {
	e := echo.New()

	NewStartWSHandlers(e.Group("/send", middleware.JWTMiddleware), sendKafka, userLogic)

	return e
}
