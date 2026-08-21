package service

import (
	"context"

	"github.com/gamee1910/volt/internal/interfaces/http/payload/request"
	"github.com/gamee1910/volt/internal/interfaces/http/payload/response"
)

type ElectricityService interface {
	LoginEVN(ctx context.Context, username, password string) error
	FetchAndSyncMonthlyUsage(ctx context.Context, req request.DailyPowerUsageRequest) (*response.MonthlyElectricityResponse, error)
	GetAll(ctx context.Context) (*response.MonthlyElectricityResponse, error)
}
