package log

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name             string
		level            string
		format           string
		expectedLoggerFn func(*testing.T, *logrus.Logger)
		expectedErrFn    assert.ErrorAssertionFunc
	}{
		{
			name:   "return error for invalid level",
			level:  "unknown",
			format: FormatJSON,
			expectedLoggerFn: func(t *testing.T, l *logrus.Logger) {
				assert.Nil(t, l)
			},
			expectedErrFn: assert.Error,
		},
		{
			name:   "return error for invalid format",
			level:  logrus.InfoLevel.String(),
			format: "unknown",
			expectedLoggerFn: func(t *testing.T, l *logrus.Logger) {
				assert.Nil(t, l)
			},
			expectedErrFn: assert.Error,
		},
		{
			name:   "return logger with json formatter",
			level:  logrus.InfoLevel.String(),
			format: FormatJSON,
			expectedLoggerFn: func(t *testing.T, l *logrus.Logger) {
				assert.Equal(t, logrus.InfoLevel, l.Level)
				assert.Equal(
					t,
					&logrus.JSONFormatter{
						TimestampFormat: time.RFC3339Nano,
					},
					l.Formatter,
				)
			},
			expectedErrFn: assert.NoError,
		},
		{
			name:   "return logger with text formatter",
			level:  logrus.InfoLevel.String(),
			format: FormatText,
			expectedLoggerFn: func(t *testing.T, l *logrus.Logger) {
				assert.Equal(t, logrus.InfoLevel, l.Level)
				assert.Equal(
					t,
					&logrus.TextFormatter{
						TimestampFormat: time.RFC3339Nano,
					},
					l.Formatter,
				)
			},
			expectedErrFn: assert.NoError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger, err := New(test.level, test.format)
			test.expectedErrFn(t, err)
			test.expectedLoggerFn(t, logger)
		})
	}
}

func TestContext(t *testing.T) {
	l := logrus.NewEntry(logrus.New())
	ctx := context.Background()

	ctx = ContextWith(ctx, l)
	ctxL := FromContext(ctx)

	assert.Equal(t, l, ctxL)
}
