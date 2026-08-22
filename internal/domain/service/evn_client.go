package service

import (
	"context"

	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/response"
)

type EVNClient interface {
	Login(ctx context.Context, username string, password string) error

	GetDailyPowerUsageData(ctx context.Context, req request.DailyPowerUsageRequest) (*response.DailyPowerUsageResponse, error)
}
