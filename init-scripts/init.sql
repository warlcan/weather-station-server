CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

CREATE TABLE Users (
    UID UUID PRIMARY KEY,
    Email VARCHAR(255) NOT NULL UNIQUE,
    ApiTokenHash VARCHAR(255) NOT NULL
);

CREATE TABLE Gateways (
    UID UUID PRIMARY KEY,
    GatewayHardwareID VARCHAR(255) NOT NULL UNIQUE,
    UserUID UUID NOT NULL REFERENCES Users(UID) ON DELETE CASCADE
);

CREATE TABLE WeatherStations (
    UID BIGINT PRIMARY KEY,
    WeatherStationHardwareID VARCHAR(100) NOT NULL UNIQUE,
    GatewayUID UUID NOT NULL REFERENCES Gateways(UID) ON DELETE CASCADE,
    ErrorMask INTEGER DEFAULT 0,
    Name VARCHAR(100)
);

CREATE TABLE Measurements(
    Time TIMESTAMPTZ NOT NULL,
    WeatherStationUID BIGINT NOT NULL REFERENCES WeatherStations(UID) ON DELETE CASCADE,
    Temperature NUMERIC(5, 2),
    Humidity NUMERIC(5, 2),
    Pressure NUMERIC(6, 2),
    VoltageLevel NUMERIC(4, 2),
    PRIMARY KEY (Time, WeatherStationUID)
);


SELECT create_hypertable('Measurements', 'time');