package decorator

import (
	"context"
	"fmt"

	"github.com/gamee1910/volt/pkg/logger"
)

type commandLoggingDecorator[C any] struct {
	base   CommandHandler[C]
	logger *logger.Logger
}

func (d commandLoggingDecorator[C]) Handle(
	ctx context.Context,
	cmd C,
) (err error) {
	action := generateActionName(cmd)

	d.logger.Info(
		"handle command",
		"command", action,
		"body", fmt.Sprintf("%#v", cmd),
	)

	d.logger.Debug("executing command")

	defer func() {
		if err != nil {
			d.logger.Error(
				"command execution failed",
				"command", action,
				"error", err,
			)
			return
		}

		d.logger.Info(
			"command executed successfully",
			"command", action,
		)
	}()

	return d.base.Handle(ctx, cmd)
}

type queryLoggingDecorator[Q any, R any] struct {
	base   QueryHandler[Q, R]
	logger *logger.Logger
}

func (d queryLoggingDecorator[Q, R]) Handle(
	ctx context.Context,
	q Q,
) (result R, err error) {
	action := generateActionName(q)

	d.logger.Info(
		"handle query",
		"query", action,
		"body", fmt.Sprintf("%#v", q),
	)

	d.logger.Debug("executing query")

	defer func() {
		if err != nil {
			d.logger.Error(
				"query execution failed",
				"query", action,
				"error", err,
			)
			return
		}

		d.logger.Info(
			"query executed successfully",
			"query", action,
		)
	}()

	return d.base.Handle(ctx, q)
}
