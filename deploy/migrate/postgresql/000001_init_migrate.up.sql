CREATE TABLE
    IF NOT EXISTS sensor_data (
        time TIMESTAMPTZ NOT NULL,
        device_id TEXT NOT NULL,
        temperature DOUBLE PRECISION NULL,
        humidity DOUBLE PRECISION NULL
    );