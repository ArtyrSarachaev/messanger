package http

import (
	"messanger/internal/entity"
	"messanger/internal/middleware"

	"github.com/labstack/echo/v4"
)

type httpServer struct {
	userLogic entity.UserLogic
}

func StartHttpServer(userLogic entity.UserLogic) *echo.Echo {
	e := echo.New()

	NewAuth(e.Group("/"), userLogic)

	NewUserHandler(e.Group("/user",
		middleware.JWTMiddleware,
		middleware.AddLoggerInHandlerContext),
		userLogic)

	return e
}
