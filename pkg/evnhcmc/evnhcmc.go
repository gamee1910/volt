package evnhcmc

import (
	"context"
)

type EVNClient interface {
	GetDailyPowerUsageData(ctx context.Context, req DailyPowerUsageRequest) (*DailyPowerUsageResponse, error)
}
