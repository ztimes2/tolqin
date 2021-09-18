package log

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	formatJSON = "json"
	formatText = "text"
)

func New(level, format string) (*logrus.Logger, error) {
	logger := logrus.New()

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(lvl)

	var formatter logrus.Formatter
	switch format {
	case formatJSON:
		formatter = &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		}
	case formatText:
		formatter = &logrus.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		}
	default:
		return nil, fmt.Errorf("invalid log format: %q", format)
	}
	logger.SetFormatter(formatter)

	return logger, nil
}

type contextKey struct{}

var keyLogEntry contextKey = struct{}{}

func ContextWith(ctx context.Context, l *logrus.Entry) context.Context {
	return context.WithValue(ctx, keyLogEntry, l)
}

func FromContext(ctx context.Context) *logrus.Entry {
	l, _ := ctx.Value(keyLogEntry).(*logrus.Entry)
	return l
}
