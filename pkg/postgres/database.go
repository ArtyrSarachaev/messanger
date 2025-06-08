package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"messanger/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pkg/errors"
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

func NewDB(ctx context.Context, config config.Config) (*sql.DB, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Postgres.UserName,
		config.Postgres.Password,
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.NameDB)

	db, err := sql.Open("pgx", connString)
	if err != nil {
		return nil, errors.Wrap(err, "cant open connection for postgres")
	}

	return db, nil
}
