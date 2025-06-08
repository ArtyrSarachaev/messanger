package http

import (
	"messanger/internal/entity"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userLogic entity.UserLogic
}

func NewUserHandler(rg *echo.Group, userLogic entity.UserLogic) {
	u := UserHandler{userLogic: userLogic}
	rg.GET("/", u.getUser)
}

func (u *UserHandler) getUser(c echo.Context) error {
	ctx := c.Request().Context()

	username := c.QueryParam("username")
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	respUser, err := u.userLogic.ByFullName(ctx, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if respUser.Username == "" {
		return c.JSON(http.StatusOK, "user is not exist")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"username": respUser.Username,
	})
}
