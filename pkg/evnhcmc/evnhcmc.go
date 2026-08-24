package evnhcmc

import (
	"context"
)

type EVNClient interface {
	Login(ctx context.Context, username string, password string) error
	GetDailyPowerUsageData(ctx context.Context, req DailyPowerUsageRequest) (*DailyPowerUsageResponse, error)
}
