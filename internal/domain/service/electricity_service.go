package service

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
)

type ElectricityService interface {
	Save(context.Context, *entity.ElectricityConsumption) error
	GetAll(context.Context) ([]*entity.ElectricityConsumption, error)
	GetByDate(context.Context, time.Time) (*entity.ElectricityConsumption, error)
}
