package service

import (
	"context"

	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/response"
)

type ElectricityService interface {
	DailyPowerUsage(ctx context.Context, req request.GetUsageRequest) error
	GetAll(ctx context.Context) (*response.ElectricityResponse, error)
	GetYesterDayUsage(ctx context.Context) (*response.ElectricityConsumptionResponse, error)
}
