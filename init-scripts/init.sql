CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

CREATE TABLE users (
    uid UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE gateways (
    uid UUID PRIMARY KEY,
    gateway_hardware_id UUID NOT NULL UNIQUE,
    user_uid UUID NOT NULL REFERENCES users(uid) ON DELETE CASCADE
);

CREATE TABLE weather_stations (
    uid BIGSERIAL PRIMARY KEY,
    weather_station_hardware_id BIGINT NOT NULL UNIQUE,
    gateway_uid UUID NOT NULL REFERENCES gateways(uid) ON DELETE CASCADE,
    error_mask SMALLINT DEFAULT 0,
    name VARCHAR(100)
);

CREATE TABLE measurements(
    time TIMESTAMPTZ NOT NULL,
    weather_station_uid BIGINT NOT NULL REFERENCES weather_stations(uid),
    temperature REAL,
    humidity REAL,
    pressure REAL,
    voltage_level REAL,
    PRIMARY KEY (time, weather_station_uid)
);


SELECT create_hypertable('measurements', 'time');

ALTER TABLE measurements SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'weather_station_uid',
    timescaledb.compress_orderby = 'time ASC'
);
SELECT add_compression_policy('measurements', INTERVAL '7 days');