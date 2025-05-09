package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"messanger/internal/entity"
	"messanger/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	cantExecQueryWithError = "cant exec query '%s', with error: %v"
)

type poolDB struct {
	Pool *pgxpool.Pool
}

func NewPool(ctx context.Context, config config.Config) (poolDB /*Возвращать приватную структуру из пакета - плохая практика*/, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=10",
		config.Postgres.UserName,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.NameDB)

	// configConnect -> connectConfig так правильнее
	configConnect, err := pgxpool.ParseConfig(connString)
	if err != nil {
		// Для вставки ошибки в fmt.Errorf() используется %w, а не %v.
		// Тогда ошибку можно будет проверить через errors.Is()
		return poolDB{}, fmt.Errorf("failed to parse conn string: %v", err)
	}

	// pgxpool.ParseConfig() парсит строчку вместе с этими параметрами
	// и лучше все эти константы вынести в конфиг
	configConnect.MaxConns = 10
	configConnect.MinConns = 2
	configConnect.MaxConnLifetime = 30 * time.Minute
	configConnect.MaxConnIdleTime = 10 * time.Minute
	configConnect.HealthCheckPeriod = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, configConnect)
	if err != nil {
		return poolDB{}, fmt.Errorf("failed to create connection pool: %v", err)
	}

	// Ещё раз декларируешь err. Правильнее сделать "=" вместо ":="
	if err := pool.Ping(ctx); err != nil {
		return poolDB{}, fmt.Errorf("failed to ping database: %v", err)
	}

	return poolDB{Pool: pool}, nil
}

// Коммент ненужный удаляй или помечай TODO или FIXME
// Обязательно с пояснением

/*
	func New(config config.Config) *database {
		postgres := config.Postgres
		connectStringToDB := postgres.СreateConnectionString()
		db, err := sql.Open("postgres", connectStringToDB)
		if err != nil {
			panic(err)
		}

		db.SetMaxOpenConns(int(postgres.MaxOpenConns))
		db.SetMaxIdleConns(int(postgres.MaxIdleConns))
		db.SetConnMaxLifetime(time.Duration(postgres.LifetimeConn) * time.Minute)

		return &database{DB: db}
	}
*/

// Поясняющий комментарий в глобальной области пакета должен начинаться с названия метода, структуры или
// переменной, которую ты описываешь

// CheckIsTableExists проверяем существует таблица или нет. Костыль для тестового задания
func (p *poolDB) CheckIsTableExists(ctx context.Context) (bool, error) {
	var result bool
	query := fmt.Sprintf(`SELECT EXISTS (
            SELECT 1
            FROM information_schema.tables
            WHERE table_schema = 'public'
            AND (table_name = '%s'
			OR table_name = '%s'))`,
		entity.MessagesTable, entity.UsersTable)
	err := p.Pool.QueryRow(ctx, query).Scan(&result)
	if err != nil {
		return false, fmt.Errorf(cantExecQueryWithError, query, err)
	}
	return result, nil
}

func (p *poolDB) CreateTables(ctx context.Context, pathToCreateDatabase string) error {
	query, err := os.ReadFile(pathToCreateDatabase)
	if err != nil {
		return fmt.Errorf("cant read from path '%s', with error %v", pathToCreateDatabase, err)
	}

	_, err = p.Pool.Exec(ctx, string(query))
	if err != nil {
		return fmt.Errorf(cantExecQueryWithError, pathToCreateDatabase, err)
	}

	return nil
}

/**

Короче, то, что ты в методах CheckIsTableExists и CreateTables пытаешься велосипед изобрести, это здорово.
Но для этого есть специальных механизм - называется миграции. Написал в comments.md в корне
*/
