package entity

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidInput       = errors.New("invalid input")
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// Credentials — тело запросов /register и /login.
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLogic interface {
	Register(ctx context.Context, creds Credentials) (User, error)
	Login(ctx context.Context, creds Credentials) (string, error)
	ByUsername(ctx context.Context, username string) (User, error)
	ByID(ctx context.Context, userID string) (User, error)
}

type UserRepository interface {
	Create(ctx context.Context, username, passwordHash string) (User, error)
	ByUsername(ctx context.Context, username string) (User, error)
	ByID(ctx context.Context, userID string) (User, error)
}

type TokenManager interface {
	Generate(userID string) (string, error)
	Parse(token string) (string, error)
}
