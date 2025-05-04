package http

import (
	"messanger/internal/entity"
	"net/http"

	"github.com/labstack/echo/v4"
)

type httpAuth struct {
	userLogic entity.UserLogic
}

func NewAuth(rg *echo.Group, userLogic entity.UserLogic) {
	h := httpAuth{userLogic: userLogic}

	rg.POST("login", h.login)
	rg.POST("register", h.register)
}

func (h *httpAuth) register(c echo.Context) error {
	ctx := c.Request().Context()

	var login entity.User
	if err := c.Bind(&login); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.userLogic.Register(ctx, login)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, "user register successfully")
}

func (h *httpAuth) login(c echo.Context) error {
	ctx := c.Request().Context()
	var user entity.User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.userLogic.Login(ctx, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}
