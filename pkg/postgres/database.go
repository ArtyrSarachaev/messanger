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

func NewPool(ctx context.Context, config config.Config) (poolDB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=10",
		config.Postgres.UserName,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.NameDB)

	configConnect, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return poolDB{}, fmt.Errorf("failed to parse conn string: %v", err)
	}

	configConnect.MaxConns = 10
	configConnect.MinConns = 2
	configConnect.MaxConnLifetime = 30 * time.Minute
	configConnect.MaxConnIdleTime = 10 * time.Minute
	configConnect.HealthCheckPeriod = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, configConnect)
	if err != nil {
		return poolDB{}, fmt.Errorf("failed to create connection pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return poolDB{}, fmt.Errorf("failed to ping database: %v", err)
	}

	return poolDB{Pool: pool}, nil
}

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
// проверяем существуют таблица или нет. Костыль для тестового задания
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
