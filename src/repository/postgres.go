package repository

import (
	"context"
	"fmt"
	"os"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PostgresDB *PostgresRepository

type PostgresRepository struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func InitPostgres() error {
	ctx := context.Background()

	dsn := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s",
		os.Getenv("PG_DB_USER"),
		os.Getenv("PG_DB_PASSWORD"),
		os.Getenv("PG_DB_PORT"),
		os.Getenv("PG_DB_NAME"),
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("Postgres pool creation failed: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("Postgres ping failed: %w", err)
	}

	PostgresDB = &PostgresRepository{
		pool: pool,
		ctx:  ctx,
	}

	return nil
}

func (r *PostgresRepository) CreateUser(email string) (string, error) {
	query := "INSERT INTO users (uid, email) VALUES ($1, $2)"
	uid := uuid.New().String()
	_, err := r.pool.Exec(r.ctx, query, uid, email)
	return uid, err
}

func (r *PostgresRepository) GetUserId(email string) (string, error) {
	query := "SELECT uid FROM users WHERE email = $1"
	var uid string
	err := r.pool.QueryRow(r.ctx, query, email).Scan(&uid)
	return uid, err
}
