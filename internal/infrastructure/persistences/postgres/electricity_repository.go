package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gamee1910/volt/internal/domain"
	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
)

type ElectricityRepository struct {
	db *sql.DB
}

func NewElectricityRepository(db *sql.DB) repository.ElectricityRepository {
	return &ElectricityRepository{db: db}
}

func (repository *ElectricityRepository) Insert(
	ctx context.Context, req *entity.ElectricityConsumption,
) error {
	query := `
		INSERT INTO electricity_consumption(reading_date, consumption_kwh, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (reading_date) DO UPDATE
		SET consumption_kwh = EXCLUDED.consumption_kwh
	`

	_, err := repository.db.ExecContext(ctx, query, req.ReadingDate, req.ConsumptionKWh)
	if err != nil {
		return fmt.Errorf("insert electricity consumption: %w", err)
	}

	return nil
}

func (repository *ElectricityRepository) FindByDate(
	ctx context.Context, date time.Time,
) (*entity.ElectricityConsumption, error) {
	query := `SELECT id, reading_date, consumption_kwh, created_at FROM electricity_consumption WHERE reading_date = $1`

	var result entity.ElectricityConsumption

	err := repository.db.QueryRowContext(ctx, query, date).Scan(
		&result.ID, &result.ReadingDate, &result.ConsumptionKWh, &result.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrElectricityConsumptionNotFound
		}
		return nil, fmt.Errorf("find by date: %w", err)
	}

	return &result, nil
}
func (repository *ElectricityRepository) FindAll(
	ctx context.Context,
) ([]*entity.ElectricityConsumption, error) {

	query := `SELECT id, reading_date, consumption_kwh, created_at FROM electricity_consumption`

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
			&consumption.ReadingDate,
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
