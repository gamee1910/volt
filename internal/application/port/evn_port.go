package port

import (
	"context"

	"github.com/gamee1910/volt/internal/application/dto"
)

type EVNClient interface {
	Login(ctx context.Context, username string, password string) error
	GetDailyPowerUsageData(ctx context.Context, req dto.DailyPowerUsageRequest) (*dto.DailyPowerUsageResponse, error)
}
