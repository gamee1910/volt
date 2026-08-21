ALTER TABLE electricity_consumption
    DROP CONSTRAINT electricity_consumption_measurement_date_key;

ALTER TABLE electricity_consumption
    RENAME COLUMN measurement_date TO reading_date;

ALTER TABLE electricity_consumption
    ADD CONSTRAINT electricity_consumption_reading_date_key
        UNIQUE (reading_date);