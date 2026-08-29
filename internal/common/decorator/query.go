package decorator

import (
	"context"

	"github.com/gamee1910/volt/pkg/logger"
)

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

func ApplyQueryDecorators[H any, R any](
	handler QueryHandler[H, R],
	logger *logger.Logger,
	metricClient MetricClient,
) QueryHandler[H, R] {
	return queryLoggingDecorator[H, R]{
		base: queryMetricsDecorator[H, R]{
			base:   handler,
			client: metricClient,
		},
		logger: logger,
	}
}
