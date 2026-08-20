package client

import (
	"context"

	"github.com/gamee1910/volt/internal/interfaces/http/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/response"
)

type EvnhcmcClient interface {
	GetDailyPowerUsageData(context.Context, request.DailyPowerUsageRequest) (*response.DailyPowerUsageResponse, error)
	Login(ctx context.Context) error
}
