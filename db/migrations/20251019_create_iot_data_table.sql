-- +goose Up
CREATE TABLE iot_data (
    time        TIMESTAMPTZ NOT NULL,
    device_id   TEXT NOT NULL,
    type        TEXT NOT NULL,  -- 'electricity' of 'gas'
    data        JSONB NOT NULL,
    PRIMARY KEY (time, device_id, type)
);

SELECT create_hypertable('iot_data', 'time');
-- +goose Down
DROP TABLE IF EXISTS iot_data;
