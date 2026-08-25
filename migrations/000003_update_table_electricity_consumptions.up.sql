ALTER TABLE electricity_consumption
    ADD COLUMN updated_at TIMESTAMPTZ;

UPDATE electricity_consumption
SET updated_at = NOW()
WHERE updated_at IS NULL;

ALTER TABLE electricity_consumption
    ALTER COLUMN updated_at SET NOT NULL;