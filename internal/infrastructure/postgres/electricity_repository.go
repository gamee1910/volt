package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
)

type ElectricityRepository struct {
	db *sql.DB
}

func NewElectricityRepository(db *sql.DB) repository.ElectricityRepository {
	return &ElectricityRepository{
		db: db,
	}
}

func (r *ElectricityRepository) Save(
	ctx context.Context,
	data *entity.ElectricityConsumption,
) error {
	panic("unimplemented")
	return nil
}

func (r *ElectricityRepository) FindByDate(
	ctx context.Context,
	date time.Time,
) (*entity.ElectricityConsumption, error) {
	panic("unimplemented")
	return nil, nil
}
