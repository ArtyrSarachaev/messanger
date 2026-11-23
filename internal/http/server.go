package http

import (
	"errors"
	"net/http"
	"strconv"

	"messanger/internal/entity"
	"messanger/internal/middleware"

	"github.com/labstack/echo/v4"
)

type handler struct {
	users    entity.UserLogic
	messages entity.MessageLogic
}

func NewServer(users entity.UserLogic, messages entity.MessageLogic, tokens entity.TokenManager) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	h := handler{users: users, messages: messages}

	e.GET("/health", h.health)
	e.POST("/register", h.register)
	e.POST("/login", h.login)

	auth := middleware.JWT(tokens)
	e.GET("/me", h.me, auth)
	e.GET("/users/:username", h.getUser, auth)
	e.GET("/messages/:username", h.getDialog, auth)

	return e
}

func (h *handler) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) register(c echo.Context) error {
	var creds entity.Credentials
	if err := c.Bind(&creds); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	user, err := h.users.Register(c.Request().Context(), creds)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *handler) login(c echo.Context) error {
	var creds entity.Credentials
	if err := c.Bind(&creds); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	token, err := h.users.Login(c.Request().Context(), creds)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *handler) me(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := entity.UserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	user, err := h.users.ByID(ctx, userID)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *handler) getUser(c echo.Context) error {
	user, err := h.users.ByUsername(c.Request().Context(), c.Param("username"))
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(http.StatusOK, user)
}

func (h *handler) getDialog(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := entity.UserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	limit, err := intQueryParam(c, "limit")
	if err != nil {
		return err
	}
	offset, err := intQueryParam(c, "offset")
	if err != nil {
		return err
	}

	messages, err := h.messages.Dialog(ctx, userID, c.Param("username"), limit, offset)
	if err != nil {
		return toHTTPError(err)
	}

	return c.JSON(http.StatusOK, messages)
}

func intQueryParam(c echo.Context, name string) (int, error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return 0, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, name+" must be an integer")
	}

	return value, nil
}

// toHTTPError переводит доменные ошибки в HTTP-статусы.
// Внутренние ошибки не раскрываются клиенту.
func toHTTPError(err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidInput):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, entity.ErrInvalidCredentials):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, entity.ErrUserNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, entity.ErrUserAlreadyExists):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(err)
	}
}
