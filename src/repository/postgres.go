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

var (
	ErrGatewayNotFound     = errors.New("gateway not found")
	ErrStationAlreadyBound = errors.New("station already bound to another gateway")
)

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

func (r *PostgresRepository) GetGatewayByHardwareId(gatewayHardwareId string) (string, error) {
	const query = `SELECT uid FROM gateways WHERE gateway_hardware_id = $1`

	var uid string
	err := r.pool.QueryRow(r.ctx, query, gatewayHardwareId).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrGatewayNotFound
	}
	if err != nil {
		return "", err
	}
	return uid, nil
}

func (r *PostgresRepository) UpsertStationsForGateway(gatewayUid string, stationIds []int64) (int, error) {
	if len(stationIds) == 0 {
		return 0, nil
	}

	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(r.ctx)

	const query = `
		INSERT INTO weather_stations (weather_station_hardware_id, gateway_uid)
		VALUES ($1, $2)
		ON CONFLICT (weather_station_hardware_id) DO UPDATE
			SET gateway_uid = EXCLUDED.gateway_uid
			WHERE weather_stations.gateway_uid = EXCLUDED.gateway_uid
		RETURNING (xmax = 0) AS inserted;
	`

	added := 0
	for _, sid := range stationIds {
		var inserted bool
		err := tx.QueryRow(r.ctx, query, sid, gatewayUid).Scan(&inserted)
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrStationAlreadyBound
		}
		if err != nil {
			return 0, err
		}
		if inserted {
			added++
		}
	}

	if err := tx.Commit(r.ctx); err != nil {
		return 0, err
	}
	return added, nil
}
