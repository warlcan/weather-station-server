package repository

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var PostgresDB *PostgresRepository

var ErrGatewayAlreadyBound = errors.New("gateway already bound to another user")

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

func (r *PostgresRepository) UpsertUser(email string) (string, error) {
	query := `
		INSERT INTO users (uid, email) 
		VALUES (gen_random_uuid(), $1)
		ON CONFLICT (email) 
		DO UPDATE SET email = EXCLUDED.email
		RETURNING uid;
	`

	var uid string
	err := r.pool.QueryRow(r.ctx, query, email).Scan(&uid)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (r *PostgresRepository) UpsertGateway(gatewayHardwareId, userUid string) (string, error) {
	const query = `
		INSERT INTO gateways (uid, gateway_hardware_id, user_uid)
		VALUES (gen_random_uuid(), $1, $2)
		ON CONFLICT (gateway_hardware_id) DO UPDATE
			SET user_uid = EXCLUDED.user_uid
			WHERE gateways.user_uid = EXCLUDED.user_uid
		RETURNING uid;
	`

	var uid string
	err := r.pool.QueryRow(r.ctx, query, gatewayHardwareId, userUid).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrGatewayAlreadyBound
	}
	if err != nil {
		return "", err
	}
	return uid, nil
}
