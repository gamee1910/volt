package service

import (
	"context"

	"github.com/gamee1910/volt/internal/interfaces/http/transport/response"
	"github.com/gamee1910/volt/pkg/evnhcm"
)

type ElectricityService interface {
	LoginEVN(ctx context.Context, username, password string) error
	FetchAndSyncMonthlyUsage(ctx context.Context, req evnhcm.DailyPowerUsageRequest) (*response.MonthlyElectricityResponse, error)
}
