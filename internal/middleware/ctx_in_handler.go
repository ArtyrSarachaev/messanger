package middleware

import (
	"messanger/pkg/logger"

	"github.com/labstack/echo/v4"
)

func AddLoggerInHandlerContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_ = logger.New(c.Request().Context())

		return next(c)
	}
}
