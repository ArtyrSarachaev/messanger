package logic

import (
	"context"
	"fmt"
	"messanger/internal/entity"
	"messanger/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userLogic struct {
	userRepository entity.UserRepository
}

func NewUserLogic(userRepository entity.UserRepository) entity.UserLogic {
	return &userLogic{userRepository: userRepository}
}

func (u *userLogic) GetUserByFullName(ctx context.Context, userName string) (entity.User, error) {
	user, err := u.userRepository.GetUserByName(ctx, userName)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (u *userLogic) Login(ctx context.Context, user entity.User) (string, error) {
	log := logger.LoggerFromContext(ctx)
	userFromDb, err := u.userRepository.GetUserByName(ctx, user.Username)
	if err != nil {
		return "", err
	}
	if userFromDb.Username == "" {
		msg := fmt.Sprintf("user %s is not exist", user.Username)
		log.Info(msg)
		return "", nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromDb.Password), []byte(user.Password))
	if err != nil {
		return "", fmt.Errorf("failed to compare hash password, with error %v", err)
	}

	claims := &entity.Claims{
		UserID: userFromDb.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(entity.TokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(entity.JwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token, with error %v", err)
	}

	return tokenString, nil
}

func (u *userLogic) Register(ctx context.Context, login entity.User) error {
	log := logger.LoggerFromContext(ctx)
	user, err := u.userRepository.GetUserByName(ctx, login.Username)
	if err != nil {
		return err
	}
	if user.Username != "" {
		msg := fmt.Sprintf("user %s is already exist", login.Username)
		log.Info(msg)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password, with error %v", err)
	}
	login.Password = string(hashedPassword)

	err = u.userRepository.SaveUser(ctx, login)
	if err != nil {
		return err
	}

	return nil
}
