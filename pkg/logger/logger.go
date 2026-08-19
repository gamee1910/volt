package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *Logger
	once         sync.Once
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(env string) *Logger {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()

		cfg.Encoding = "console"
		cfg.DisableCaller = true
		cfg.DisableStacktrace = true

		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

func Init(env string) *Logger {
	once.Do(func() {
		globalLogger = NewLogger(env)
	})
	return globalLogger
}

func L() *Logger {
	if globalLogger == nil {
		Init("development")
	}
	return globalLogger
}

func Info(msg string, keysAndValues ...any) {
	L().Infow(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...any) {
	L().Errorw(msg, keysAndValues...)
}

func Infof(format string, args ...any) {
	L().SugaredLogger.Infof(format, args...)
}

func Fatal(args ...any) {
	L().SugaredLogger.Fatal(args...)
}
