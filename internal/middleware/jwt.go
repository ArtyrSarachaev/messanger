package middleware

import (
	"net/http"
	"strings"

	"messanger/internal/entity"

	"github.com/labstack/echo/v4"
)

const tokenQueryParam = "token"

// JWT проверяет токен и кладёт ID пользователя в контекст запроса.
// Токен берётся из заголовка "Authorization: Bearer <token>",
// а если его нет — из query-параметра ?token= (браузерный WebSocket не умеет
// передавать заголовки).
func JWT(tokens entity.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, ok := extractToken(c.Request())
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			userID, err := tokens.Parse(tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			ctx := entity.WithUserID(c.Request().Context(), userID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func extractToken(r *http.Request) (string, bool) {
	if header := r.Header.Get(echo.HeaderAuthorization); header != "" {
		token, found := strings.CutPrefix(header, "Bearer ")
		if !found || token == "" {
			return "", false
		}
		return token, true
	}

	if token := r.URL.Query().Get(tokenQueryParam); token != "" {
		return token, true
	}

	return "", false
}
