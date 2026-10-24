package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var PostgresDB *PostgresRepository

var ErrGatewayAlreadyBound = errors.New("gateway already bound to another user")
var (
	ErrGatewayNotFound     = errors.New("gateway not found")
	ErrStationAlreadyBound = errors.New("station already bound to another gateway")
)
var ErrStationNotFound = errors.New("weather station not found for this gateway")

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

func (r *PostgresRepository) GetStationUidForGateway(gatewayUid string, stationHardwareId int64) (int64, error) {
	const query = `
		SELECT uid FROM weather_stations
		WHERE weather_station_hardware_id = $1 AND gateway_uid = $2
	`

	var uid int64
	err := r.pool.QueryRow(r.ctx, query, stationHardwareId, gatewayUid).Scan(&uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrStationNotFound
	}
	if err != nil {
		return 0, err
	}
	return uid, nil
}

type Measurement struct {
	Time              time.Time
	WeatherStationUid int64
	Temperature       float32
	Humidity          float32
	Pressure          float32
	VoltageLevel      float32
}

func (r *PostgresRepository) InsertMeasurement(m Measurement, errorMask int16) error {
	tx, err := r.pool.Begin(r.ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(r.ctx)

	const insertQ = `
		INSERT INTO measurements
			(time, weather_station_uid, temperature, humidity, pressure, voltage_level)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (time, weather_station_uid) DO UPDATE
			SET temperature   = EXCLUDED.temperature,
			    humidity      = EXCLUDED.humidity,
			    pressure      = EXCLUDED.pressure,
			    voltage_level = EXCLUDED.voltage_level;
	`

	if _, err := tx.Exec(r.ctx, insertQ,
		m.Time, m.WeatherStationUid,
		m.Temperature, m.Humidity, m.Pressure, m.VoltageLevel,
	); err != nil {
		return err
	}

	const updateQ = `UPDATE weather_stations SET error_mask = $1 WHERE uid = $2`
	if _, err := tx.Exec(r.ctx, updateQ, errorMask, m.WeatherStationUid); err != nil {
		return err
	}

	return tx.Commit(r.ctx)
}

type AppWeatherStationRow struct {
	Uid        int64
	HardwareId int64
	Name       *string
	ErrorMask  int16
	GatewayUid string
}

type AppMeasurementRow struct {
	WeatherStationUid int64
	Time              time.Time
	Temperature       *float32
	Humidity          *float32
	Pressure          *float32
	VoltageLevel      *float32
}

func (r *PostgresRepository) ListUserStations(userUid string) ([]AppWeatherStationRow, error) {
	const q = `
		SELECT ws.uid, ws.weather_station_hardware_id, ws.name, ws.error_mask, ws.gateway_uid
		FROM weather_stations ws
		JOIN gateways g ON g.uid = ws.gateway_uid
		WHERE g.user_uid = $1
		ORDER BY ws.uid
	`
	rows, err := r.pool.Query(r.ctx, q, userUid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppWeatherStationRow
	for rows.Next() {
		var s AppWeatherStationRow
		if err := rows.Scan(&s.Uid, &s.HardwareId, &s.Name, &s.ErrorMask, &s.GatewayUid); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ListHourlyMeasurements(stationUids []int64, since time.Time) ([]AppMeasurementRow, error) {
	const q = `
		SELECT DISTINCT ON (weather_station_uid, date_trunc('hour', time))
			weather_station_uid, time, temperature, humidity, pressure, voltage_level
		FROM measurements
		WHERE weather_station_uid = ANY($1) AND time >= $2
		ORDER BY weather_station_uid, date_trunc('hour', time), time DESC
	`
	rows, err := r.pool.Query(r.ctx, q, stationUids, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppMeasurementRow
	for rows.Next() {
		var m AppMeasurementRow
		if err := rows.Scan(&m.WeatherStationUid, &m.Time, &m.Temperature, &m.Humidity, &m.Pressure, &m.VoltageLevel); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
