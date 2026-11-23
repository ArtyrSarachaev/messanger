package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"messanger/internal/entity"

	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLen = 3
	maxUsernameLen = 32
	minPasswordLen = 6
	// bcrypt учитывает только первые 72 байта пароля.
	maxPasswordBytes = 72
)

type userLogic struct {
	repo   entity.UserRepository
	tokens entity.TokenManager
}

func NewUserLogic(repo entity.UserRepository, tokens entity.TokenManager) entity.UserLogic {
	return &userLogic{repo: repo, tokens: tokens}
}

func (u *userLogic) Register(ctx context.Context, creds entity.Credentials) (entity.User, error) {
	creds.Username = strings.TrimSpace(creds.Username)
	if err := validateCredentials(creds); err != nil {
		return entity.User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return entity.User{}, fmt.Errorf("hash password: %w", err)
	}

	// Уникальность username гарантирует UNIQUE-индекс в БД:
	// репозиторий вернёт ErrUserAlreadyExists, гонки между проверкой и вставкой нет.
	user, err := u.repo.Create(ctx, creds.Username, string(hash))
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (u *userLogic) Login(ctx context.Context, creds entity.Credentials) (string, error) {
	user, err := u.repo.ByUsername(ctx, strings.TrimSpace(creds.Username))
	if errors.Is(err, entity.ErrUserNotFound) {
		return "", entity.ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		return "", entity.ErrInvalidCredentials
	}

	token, err := u.tokens.Generate(user.ID)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return token, nil
}

func (u *userLogic) ByUsername(ctx context.Context, username string) (entity.User, error) {
	return u.repo.ByUsername(ctx, strings.TrimSpace(username))
}

func (u *userLogic) ByID(ctx context.Context, userID string) (entity.User, error) {
	return u.repo.ByID(ctx, userID)
}

func validateCredentials(creds entity.Credentials) error {
	usernameLen := utf8.RuneCountInString(creds.Username)
	if usernameLen < minUsernameLen || usernameLen > maxUsernameLen {
		return fmt.Errorf("%w: username must be %d-%d characters", entity.ErrInvalidInput, minUsernameLen, maxUsernameLen)
	}
	if strings.ContainsAny(creds.Username, " \t\n/") {
		return fmt.Errorf("%w: username must not contain spaces or '/'", entity.ErrInvalidInput)
	}
	if utf8.RuneCountInString(creds.Password) < minPasswordLen {
		return fmt.Errorf("%w: password must be at least %d characters", entity.ErrInvalidInput, minPasswordLen)
	}
	if len(creds.Password) > maxPasswordBytes {
		return fmt.Errorf("%w: password is too long", entity.ErrInvalidInput)
	}
	return nil
}
