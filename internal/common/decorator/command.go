package decorator

import (
	"context"
	"fmt"
	"strings"

	"github.com/gamee1910/volt/pkg/logger"
)

type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}

func ApplyCommandDecorators[H any](
	handler CommandHandler[H],
	logger *logger.Logger,
	metricClient MetricClient,
) CommandHandler[H] {
	return commandLoggingDecorator[H]{
		base: commandMetricsDecorator[H]{
			base:   handler,
			client: metricClient,
		},
		logger: logger,
	}
}
