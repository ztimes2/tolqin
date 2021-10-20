package log

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// The following are formats of logging output.
const (
	// FormatJSON is used for logging in a structured JSON format.
	FormatJSON = "json"
	// FormatText is used for logging in a plain text format.
	FormatText = "text"
)

// New initializes a new logger using the given level and format.
func New(level, format string) (*logrus.Logger, error) {
	logger := logrus.New()

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}

	logger.SetLevel(lvl)

	var formatter logrus.Formatter
	switch format {
	case FormatJSON:
		formatter = &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		}
	case FormatText:
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

// ContextWith returns a copy of the given context with the given log entry attached
// to it.
func ContextWith(ctx context.Context, l *logrus.Entry) context.Context {
	return context.WithValue(ctx, keyLogEntry, l)
}

// FromContext retrieves a log entry from the given context if available. If the
// context does not contain a log entry, a nil log entry is returned.
func FromContext(ctx context.Context) *logrus.Entry {
	l, _ := ctx.Value(keyLogEntry).(*logrus.Entry)
	return l
}