package log

import (
	"context"
	"os"

	"github.com/hellofresh/github-cli/pkg/formatter"
	"github.com/sirupsen/logrus"
)

type loggerKeyType int

const loggerKey loggerKeyType = iota

var logger logrus.Logger

func init() {
	logger = logrus.Logger{
		Out:       os.Stderr,
		Formatter: &formatter.CliFormatter{},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
}

// NewContext returns a context that has a logrus logger
func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, WithContext(ctx))
}

// WithContext returns a logrus logger from the context
func WithContext(ctx context.Context) *logrus.Logger {
	if ctx == nil {
		return &logger
	}

	if ctxLogger, ok := ctx.Value(loggerKey).(*logrus.Logger); ok {
		return ctxLogger
	}

	return &logger
}
