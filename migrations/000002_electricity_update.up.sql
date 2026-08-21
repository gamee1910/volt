ALTER TABLE electricity_consumption
    DROP CONSTRAINT electricity_consumption_reading_date_key;

ALTER TABLE electricity_consumption
    RENAME COLUMN reading_date TO measurement_date;

ALTER TABLE electricity_consumption
    ALTER COLUMN measurement_date SET NOT NULL;

ALTER TABLE electricity_consumption
    ADD CONSTRAINT electricity_consumption_measurement_date_key
        UNIQUE (measurement_date);