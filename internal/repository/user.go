package repository

import (
	"context"
	"fmt"
	"messanger/internal/entity"
	"strings"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	poolPG *pgxpool.Pool
}

func NewUserRepository(poolPG *pgxpool.Pool) entity.UserRepository {
	return &userRepository{poolPG: poolPG}
}

func (u *userRepository) SaveUser(ctx context.Context, userLogin entity.User) error {
	query := fmt.Sprintf(`INSERT INTO db_pg.public.%[1]s (%[2]s, %[3]s)
	VALUES ('%[4]s', '%[5]s')`,
		usersTable,
		usernameColumn,
		passwordColumn,
		userLogin.Username,
		userLogin.Password)

	_, err := u.poolPG.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf(cantExecQueryWithError, query, err)
	}
	return nil
}

func (u *userRepository) GetUserByName(ctx context.Context, userName string) (entity.User, error) {
	query := fmt.Sprintf(`SELECT %[1]s, %[2]s, %[3]s FROM db_pg.public.%[4]s WHERE %[2]s='%[5]s'`,
		idColumn,
		usernameColumn,
		passwordColumn,
		usersTable,
		userName)

	var user entity.User
	err := u.poolPG.QueryRow(ctx, query).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if strings.Compare(err.Error(), pgx.ErrNoRows.Error()) == 0 {
			return entity.User{}, nil
		}
		return entity.User{}, fmt.Errorf(cantExecQueryWithError, query, err)
	}

	return user, nil
}
