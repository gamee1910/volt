package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
)

type ElectricityRepository struct {
	db *sql.DB
}

func NewElectricityRepository(db *sql.DB) repository.ElectricityRepository {
	return &ElectricityRepository{db: db}
}

func (repository *ElectricityRepository) Upsert(ctx context.Context, req *entity.ElectricityConsumption) error {
	query := `
		INSERT INTO electricity_consumption(
			measurement_date, consumption_kwh, created_at
		) VALUES (
		          $1, $2, NOW()
		) 
	  	ON CONFLICT (measurement_date) 
		DO UPDATE SET 
			  consumption_kwh = EXCLUDED.consumption_kwh
 	`
	_, err := repository.db.ExecContext(
		ctx,
		query,
		req.MeasurementDate,
		req.ConsumptionKWh,
	)

	if err != nil {
		return fmt.Errorf("upsert electricity consumption: %w", err)
	}

	return nil
}

func (repository *ElectricityRepository) GetByDate(ctx context.Context, date time.Time) (*entity.ElectricityConsumption, error) {

	query := `
		SELECT id, measurement_date, consumption_kwh, created_at
		FROM electricity_consumption
		WHERE measurement_date = $1
	`

	var consumption entity.ElectricityConsumption

	err := repository.db.QueryRowContext(
		ctx, query, date,
	).Scan(
		&consumption.ID,
		&consumption.MeasurementDate,
		&consumption.ConsumptionKWh,
		&consumption.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("get electricity consumption by date: %w", err)
	}

	return &consumption, nil

}

func (repository *ElectricityRepository) GetAll(ctx context.Context) ([]*entity.ElectricityConsumption, error) {

	query := `SELECT id, measurement_date, consumption_kwh, created_at FROM electricity_consumption`

	rows, err := repository.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all electricity consumption: %w", err)
	}
	defer rows.Close()

	consumptions := make([]*entity.ElectricityConsumption, 0)

	for rows.Next() {
		var consumption entity.ElectricityConsumption

		err := rows.Scan(
			&consumption.ID,
			&consumption.MeasurementDate,
			&consumption.ConsumptionKWh,
			&consumption.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan electricity consumption: %w", err)
		}

		consumptions = append(consumptions, &consumption)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate electricity consumption: %w", err)
	}

	return consumptions, nil
}

func (repository *ElectricityRepository) GetTotalConsumption(ctx context.Context, fromDate, toDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(consumption_kwh), 0) AS total_consumption
		FROM electricity_consumption
		WHERE measurement_date >= $1 AND measurement_date <= $2
	`

	var totalKWh float64

	err := repository.db.QueryRowContext(
		ctx, query, fromDate, toDate,
	).Scan(&totalKWh)

	if err != nil {
		return 0, fmt.Errorf("get total consumption: %w", err)
	}
	return totalKWh, nil
}
