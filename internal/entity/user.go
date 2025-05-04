package entity

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIDKey = "user_id"
)

const (
	UsersTable    = "users"
	MessagesTable = "messages"
)

const (
	JwtSecret = "99d3fb44d62f3f0159d247efcf7d9d2247d50c5d9f9befb24c02495b86fce745edbede5862c7519eb0411ea526a87f4b0863dab889713383e0935c55a3bd6e1843419f0d438d63433b1b8a40af290b88a4a7e15fbac8db6a586490d76ce4730f725ce67fe2f9696a7b37906bb12d813d581defb16ae2d370931b5068f6cace9f9851a1e31be2cbc9948c1d6fc0c4ca6b05df609697b448ba0c9ce2083536a7f3ee740b5af982a830924eb799545c9687f5c07f618a59ba006374b1a4afa0697ed3e9628bea03f6fcea03be7baea59e617aa758b080740c9b47b53e4380e84dea025dab39c3afc9e04521b42f947bdfa3deea30a2a74e662a6f2d934693ef1305"
	TokenTTL  = 24 * time.Hour
)

type User struct {
	ID       int64  `json:"-"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type UserWS struct {
	Username  string
	IpAddress string
	Text      string
}

type UserLogic interface {
	GetUserByFullName(ctx context.Context, userName string) (User, error)
	Register(ctx context.Context, login User) error
	Login(ctx context.Context, user User) (string, error)
}

type UserRepository interface {
	SaveUser(ctx context.Context, userLogin User) error
	GetUserByName(ctx context.Context, userName string) (User, error)
}
