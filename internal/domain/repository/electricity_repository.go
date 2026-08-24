package repository

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
)

type ElectricityRepository interface {
	Upsert(ctx context.Context, req *entity.ElectricityConsumption) error
	GetByDate(ctx context.Context, date time.Time) (*entity.ElectricityConsumption, error)
	GetAll(ctx context.Context) ([]*entity.ElectricityConsumption, error)
	GetTotalConsumption(ctx context.Context, fromDate, toDate time.Time) (float64, error)
}
