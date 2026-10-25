package logger

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
)

// global holds the process-wide logger installed by the application bootstrap.
var global atomic.Pointer[Logger]

// nopLogger silently discards entries; it is used until a logger is installed.
var nopLogger = &Logger{base: zap.NewNop()}

// SetGlobal installs the process-wide logger. Passing nil restores the no-op
// logger, which is useful in tests.
func SetGlobal(l *Logger) {
	global.Store(l)
}

// L returns the process-wide logger, or a no-op logger when none is installed.
func L() (l *Logger) {
	if l = global.Load(); l != nil {
		return
	}
	l = nopLogger
	return
}

// Debug logs at debug level using the process-wide logger.
func Debug(msg string, fields ...Field) {
	L().Debug(msg, fields...)
}

// Info logs at info level using the process-wide logger.
func Info(msg string, fields ...Field) {
	L().Info(msg, fields...)
}

// Warn logs at warn level using the process-wide logger.
func Warn(msg string, fields ...Field) {
	L().Warn(msg, fields...)
}

// Error logs at error level using the process-wide logger.
func Error(msg string, fields ...Field) {
	L().Error(msg, fields...)
}

// contextKey is an unexported type so context keys cannot collide with keys
// defined by other packages.
type contextKey struct{}

// NewContext returns a child context carrying l. Request middleware installs
// the request-scoped logger once and handlers retrieve it with FromContext.
func NewContext(ctx context.Context, l *Logger) (derived context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	derived = context.WithValue(ctx, contextKey{}, l)
	return
}

// FromContext returns the logger stored in ctx, or the process-wide logger when
// ctx carries none.
func FromContext(ctx context.Context) (l *Logger) {
	if ctx != nil {
		if stored, ok := ctx.Value(contextKey{}).(*Logger); ok && stored != nil {
			l = stored
			return
		}
	}
	l = L()
	return
}
