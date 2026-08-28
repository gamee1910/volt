package service

import (
	"context"

	"github.com/gamee1910/volt/internal/application/dto"
)

type ElectricityService interface {
	DailyPowerUsage(ctx context.Context, param dto.GetUsageParam) error
	GetAll(ctx context.Context) (*dto.ElectricitySummaryDTO, error)
	GetYesterDayUsage(ctx context.Context) (*dto.ElectricityConsumptionDTO, error)
}
