package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTManager выпускает и проверяет JWT-токены (HS256).
type JWTManager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *JWTManager) Generate(userID string) (string, error) {
	now := m.now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	})

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (m *JWTManager) Parse(tokenString string) (string, error) {
	var c claims

	token, err := jwt.ParseWithClaims(tokenString, &c, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid || c.UserID == "" {
		return "", ErrInvalidToken
	}

	return c.UserID, nil
}
