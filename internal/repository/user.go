package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"messanger/internal/entity"
	"messanger/internal/repository/queries"
)

const (
	// op - сокращение от operation, так же можно называть span - saveUserSpan
	saveUserOp      = "repository.UserRepository.Save"
	getUserByNameOp = "repository.UserRepository.GetByName"
)

// Опять приватная структура
type userRepository struct {
	poolPG *pgxpool.Pool // просто pool
}

// NewUserRepository Возвращать интерфейс из конструктора это bad practice
func NewUserRepository(poolPG *pgxpool.Pool) entity.UserRepository {
	return &userRepository{poolPG: poolPG}
}

// SaveUser у тебя в названии структуры уже есть User, так что тут нет смысла дублировать - просто Save.
// Так же userLogin, метод и так называется Save зачем указывать что это login.
// В го, если понятен контекст, нужно именовать переменные короче. Типа у тебя UserRepository
// и если ты назовёшь переменную просто u, то из контекста будет понятно, что это User,
// но тут всё равно нужно чувствовать баланс и не опускать слишком важные детали.
// Вот советую прочитать https://go.dev/doc/effective_go. Если впадлу англ читать и яндексом пользуешься
// то врубай переводчик, он норм переводит
// Вообще советую на этой страничке(https://go.dev/doc/) зависнуть, там много полезных вещей
func (u *userRepository) SaveUser(ctx context.Context, userLogin entity.User) error {
	// Никогда не используй fmt.Sprintf для запросов в БД. Так легко можно допустить SQL-инъекции(https://habr.com/ru/articles/725134/)
	// Так же для запросов в стандартной либе есть очень удобная штука - пакет embed

	// прям сюда параметры прокидывашь, а либа уже проверит на SQL-инъекции
	_, err := u.poolPG.Exec(ctx, queries.InsertUser, userLogin.Username, userLogin.Password)
	if err != nil {
		// и в ошибке сразу будет видно, что произошло и где "repository.UserRepository.Save: unexpected EOF"
		// если доп уточнение нужно, то можно так fmt.Errorf("%s: exec query: %w", saveUserOp, err)
		return fmt.Errorf("%s: %w", saveUserOp, err)
	}
	return nil
}

// GetUserByName -> GetByName
func (u *userRepository) GetUserByName(ctx context.Context, userName string /* тут userName, но entity.User.Username без заглавной N - надо в едином стиле(вообще правильно username) */) (entity.User, error) {
	// тут так же через embed переписать
	// ещё имя БД(db_pg) тоже обычно в запросах не пишется
	query := fmt.Sprintf(`SELECT %[1]s, %[2]s, %[3]s FROM db_pg.public.%[4]s WHERE %[2]s='%[5]s'`,
		idColumn,
		usernameColumn,
		passwordColumn,
		usersTable,
		userName)

	var user entity.User
	err := u.poolPG.QueryRow(ctx, query).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		// Для проверки ошибок надо errors.Is юзать. Ещё в либе errors много интересного, погугли почитай
		if errors.Is(err, pgx.ErrNoRows) {
			// sql.ErrNoRows - ошибка из стандартной либы, она используется всегда, чтоб показать, что запрос не вернул ни одну строку
			// и чтоб в бизнес-логике ты мог сделать errors.Is(err, sql.ErrNoRows) и понять, что пользователь не найден, а не просто какая-то ошибка
			return entity.User{}, fmt.Errorf("%s: %w", getUserByNameOp, sql.ErrNoRows)
		}

		return entity.User{}, fmt.Errorf("%s: %w", getUserByNameOp, err)
	}

	return user, nil
}
