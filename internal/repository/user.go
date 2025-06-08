package repository

import (
	"context"
	"database/sql"
	"fmt"
	"messanger/internal/entity"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type userRepository struct {
	poolPG *pgxpool.Pool
}

func NewUserRepository(poolPG *pgxpool.Pool) entity.UserRepository {
	return &userRepository{poolPG: poolPG}
}

func (u *userRepository) Save(ctx context.Context, userLogin entity.User) error {
	query := fmt.Sprintf(`INSERT INTO users (username, "password", created_at) VALUES ($1, $2, $3)`)

	_, err := u.poolPG.Exec(ctx, query, userLogin.Username, userLogin.Password, userLogin.CreatedAt)
	if err != nil {
		return errors.Wrapf(err, "cant save user with username %v", userLogin.Username)
	}
	return nil
}

func (u *userRepository) ByName(ctx context.Context, username string) (entity.User, error) {
	var (
		user   entity.User
		respID uuid.UUID
	)

	query := fmt.Sprintf(`SELECT u.id, u.username, u."password", u.created_at FROM users u WHERE u.username=$1`)

	err := u.poolPG.QueryRow(ctx, query, username).Scan(&respID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, nil
		}
		return entity.User{}, errors.Wrapf(err, "cant get user by name %v", username)
	}
	user.ID = respID.String()

	return user, nil
}

func (u *userRepository) ByUserID(ctx context.Context, userID string) (entity.User, error) {
	var (
		user   entity.User
		respID uuid.UUID
	)

	query := fmt.Sprintf(`SELECT u.id, u.username, u."password", u.created_at FROM users u WHERE u.id=$1`)

	id, err := uuid.FromString(userID)
	if err != nil {
		return entity.User{}, errors.Wrapf(err, "cant parse id %v to uuid", userID)
	}
	err = u.poolPG.QueryRow(ctx, query, id).Scan(&respID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.User{}, nil
		}
		return entity.User{}, errors.Wrapf(err, "cant get user by user id %v", userID)
	}
	user.ID = respID.String()

	return user, nil

}
