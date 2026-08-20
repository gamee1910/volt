create table electricity_consumption
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    reading_date    DATE           NOT NULL,
    consumption_kwh NUMERIC(12, 3) NOT NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    UNIQUE (reading_date)
);