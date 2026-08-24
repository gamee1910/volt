package service

import (
	"context"

	"github.com/gamee1910/volt/internal/application/dto"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/response"
)

type ElectricityService interface {
	LoginEVN(ctx context.Context, username, password string) error
	FetchAndSyncMonthlyUsage(ctx context.Context, req dto.DailyPowerUsageRequest) error
	GetAll(ctx context.Context) (*response.ElectricityResponse, error)
	GetYesterDayUsage(ctx context.Context) (*response.ElectricityConsumptionResponse, error)
}
