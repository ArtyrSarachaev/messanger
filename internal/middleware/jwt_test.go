package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"messanger/internal/entity"

	"github.com/labstack/echo/v4"
)

type stubTokens struct{}

func (stubTokens) Generate(string) (string, error) { return "", nil }

func (stubTokens) Parse(token string) (string, error) {
	if token == "good" {
		return "user-1", nil
	}
	return "", errors.New("bad token")
}

func TestJWT(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		query      string
		wantStatus int
	}{
		{name: "bearer header", header: "Bearer good", wantStatus: http.StatusOK},
		{name: "query param", query: "?token=good", wantStatus: http.StatusOK},
		{name: "no token", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer bad", wantStatus: http.StatusUnauthorized},
		{name: "no bearer prefix", header: "good", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.GET("/", func(c echo.Context) error {
				userID, ok := entity.UserIDFromContext(c.Request().Context())
				if !ok || userID != "user-1" {
					return c.NoContent(http.StatusTeapot)
				}
				return c.NoContent(http.StatusOK)
			}, JWT(stubTokens{}))

			req := httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			if tt.header != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.header)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
