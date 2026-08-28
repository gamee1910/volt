package service

import (
	"context"

	"github.com/gamee1910/volt/internal/application/types"
)

type ElectricityService interface {
	DailyPowerUsage(ctx context.Context, req types.GetUsageParam) error
	GetAll(ctx context.Context) (*types.ElectricitySummary, error)
	GetYesterDayUsage(ctx context.Context) (*types.ElectricityConsumption, error)
}
