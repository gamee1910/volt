package repository

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
)

type ElectricityRepository interface {
	Save(context.Context, *entity.ElectricityConsumption) error
	FindByDate(context.Context, time.Time) (*entity.ElectricityConsumption, error)
}
