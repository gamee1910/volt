package repository

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
)

type ElectricityRepository interface {
	Insert(context.Context, *entity.ElectricityConsumption) error
	FindByDate(context.Context, time.Time) (*entity.ElectricityConsumption, error)
	FindAll(context.Context) ([]*entity.ElectricityConsumption, error)
}
