package application

import (
	"context"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
)

type electricityService struct {
	electricityRepository repository.ElectricityRepository
}

func NewElectricityService(electricityRepository repository.ElectricityRepository) service.ElectricityService {
	return &electricityService{electricityRepository: electricityRepository}
}

func (s *electricityService) Save(ctx context.Context, req *entity.ElectricityConsumption) error {
	return s.electricityRepository.Insert(ctx, req)
}

func (s *electricityService) GetAll(ctx context.Context) ([]*entity.ElectricityConsumption, error) {
	return s.electricityRepository.FindAll(ctx)
}

func (s *electricityService) GetByDate(ctx context.Context, date time.Time) (*entity.ElectricityConsumption, error) {
	return s.electricityRepository.FindByDate(ctx, date)
}
