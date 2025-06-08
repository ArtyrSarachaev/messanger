package http

import (
	"messanger/internal/entity"
	"messanger/internal/middleware"

	"github.com/labstack/echo/v4"
)

type httpServer struct {
	userLogic    entity.UserLogic
	messageLogic entity.MessageLogic
}

func Server(userLogic entity.UserLogic, messageLogic entity.MessageLogic) *echo.Echo {
	e := echo.New()

	NewAuth(e.Group("/"), userLogic)
	NewUserHandler(e.Group("/user", middleware.JWTMiddleware), userLogic)
	NewMessageHandler(e.Group("/messages", middleware.JWTMiddleware), messageLogic)

	return e
}
