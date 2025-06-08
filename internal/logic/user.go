package logic

import (
	"context"
	"fmt"
	"messanger/internal/entity"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userLogic struct {
	userRepository entity.UserRepository
}

func NewUserLogic(userRepository entity.UserRepository) entity.UserLogic {
	return &userLogic{userRepository: userRepository}
}

func (u *userLogic) ByFullName(ctx context.Context, username string) (entity.User, error) {
	return u.userRepository.ByName(ctx, username)
}

func (u *userLogic) ByUserID(ctx context.Context, userID string) (entity.User, error) {
	return u.userRepository.ByUserID(ctx, userID)
}

func (u *userLogic) Login(ctx context.Context, user entity.User) (string, error) {
	log := zap.NewExample().Sugar()
	userFromDb, err := u.userRepository.ByName(ctx, user.Username)
	if err != nil {
		return "", err
	}
	if userFromDb.Username == "" {
		log.Infof("user %s is not exist", user.Username)
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
	log := zap.NewExample().Sugar()
	user, err := u.userRepository.ByName(ctx, login.Username)
	if err != nil {
		return err
	}
	if user.Username != "" {
		log.Infof("user %s is already exist", login.Username)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(login.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password, with error %v", err)
	}
	login.Password = string(hashedPassword)

	err = u.userRepository.Save(ctx, login)
	if err != nil {
		return err
	}

	return nil
}
