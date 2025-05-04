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

	var user entity.User
	if err := c.Bind(&user); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	respUser, err := u.userLogic.GetUserByFullName(ctx, user.Username)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if respUser.Username == "" {
		return c.JSON(http.StatusOK, "user is not exist")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"userName": user.Username,
	})
}
