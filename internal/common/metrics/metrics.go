package metrics

import "github.com/gamee1910/volt/pkg/logger"

type StatsClient struct {
	logger *logger.Logger
}

func NewStatsClient(logger *logger.Logger) StatsClient {
	return StatsClient{logger: logger}
}

func (s StatsClient) Inc(key string, value int) {
	if s.logger != nil {
		s.logger.Info("[metric]", "key", key, "value", value)
	}
}
