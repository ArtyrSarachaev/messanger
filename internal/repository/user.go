package repository

import (
	"context"
	"errors"
	"fmt"

	"messanger/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolationCode = "23505"

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) entity.UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) Create(ctx context.Context, username, passwordHash string) (entity.User, error) {
	const query = `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id::text, username, password, created_at`

	user, err := scanUser(r.pool.QueryRow(ctx, query, username, passwordHash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return entity.User{}, entity.ErrUserAlreadyExists
		}
		return entity.User{}, fmt.Errorf("create user %q: %w", username, err)
	}

	return user, nil
}

func (r *userRepository) ByUsername(ctx context.Context, username string) (entity.User, error) {
	const query = `
		SELECT id::text, username, password, created_at
		FROM users
		WHERE username = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by username %q: %w", username, err)
	}

	return user, nil
}

func (r *userRepository) ByID(ctx context.Context, userID string) (entity.User, error) {
	const query = `
		SELECT id::text, username, password, created_at
		FROM users
		WHERE id = $1::uuid`

	user, err := scanUser(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		var pgErr *pgconn.PgError
		// Некорректный UUID (22P02) считаем отсутствующим пользователем.
		if errors.Is(err, pgx.ErrNoRows) || (errors.As(err, &pgErr) && pgErr.Code == "22P02") {
			return entity.User{}, entity.ErrUserNotFound
		}
		return entity.User{}, fmt.Errorf("get user by id %q: %w", userID, err)
	}

	return user, nil
}

func scanUser(row pgx.Row) (entity.User, error) {
	var user entity.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	return user, err
}
