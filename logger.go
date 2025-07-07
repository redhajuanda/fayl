package fayl

import (
	"context"
	"io"

	"github.com/redhajuanda/fayl/tracer"

	"github.com/sirupsen/logrus"
)

// Logger represent common interface for logging function
type Logger interface {
	WithContext(ctx context.Context) Logger
	WithStack(err error) Logger
	WithParam(key string, value any) Logger
	WithParams(params map[string]any) Logger
	Errorf(format string, args ...any)
	Error(args ...any)
	Fatalf(format string, args ...any)
	Fatal(args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
	Warnf(format string, args ...any)
	Warn(args ...any)
	Debugf(format string, args ...any)
	Debug(args ...any)
}

type logger struct {
	*logrus.Entry
}

var logStore *logger

// New returns a new wrapper log
func newLogger(serviceName string) *logger {

	logStore = &logger{logrus.New().WithFields(logrus.Fields{"service": serviceName})}
	// if cfg.App.Debug {
	// 	logStore.Logger.SetLevel(logrus.DebugLevel)
	// } else {
	// 	logStore.Logger.SetLevel(logrus.InfoLevel)
	// }
	return logStore
}

// SetOutput sets the logger output.
func logSetOutput(output io.Writer) {
	logStore.Logger.SetOutput(output)
}

// SetFormatter sets the logger formatter.
func logSetFormatter(formatter logrus.Formatter) {
	logStore.Logger.SetFormatter(formatter)
}

// SetLevel sets the logger level.
func logSetLevel(level logrus.Level) {
	logStore.Logger.SetLevel(level)
}

// WithContext reads requestId and correlationId from context and adds to log field
func (l *logger) WithContext(ctx context.Context) Logger {

	entry := l.Entry

	if ctx != nil {
		requestID := tracer.GetRequestID(ctx)
		if requestID != "" {
			entry = entry.WithField("request_id", requestID)
		}

		correlationID := tracer.GetCorrelationID(ctx)
		if correlationID != "" {
			entry = entry.WithField("correlation_id", correlationID)
		}
	}
	return &logger{entry}
}

func (l *logger) WithStack(err error) Logger {

	stack := tracer.MarshalStack(err)
	return &logger{l.WithField("stack", stack)}
}

func (l *logger) WithParam(key string, value interface{}) Logger {

	return &logger{l.WithField(key, value)}
}

func (l *logger) WithParams(params map[string]any) Logger {
	return &logger{l.WithFields(logrus.Fields(params))}
}
