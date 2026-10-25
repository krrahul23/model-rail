// Package logger provides the structured logger used across ModelRail. It
// wraps zap and is deliberately a leaf package: it imports only the standard
// library, zap and the dependency-free constant package, so any other package
// can depend on it without creating an import cycle.
package logger

import (
	"fmt"
	"strings"

	"modelrail/internal/constant"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config controls logger construction. It is filled from config.yml by the
// application bootstrap.
type Config struct {
	Level    string // debug | info | warn | error; empty means info
	Encoding string // json | console; empty means json
}

// Logger is a thin, nil-safe wrapper around zap.Logger. It exposes only the
// operations the application uses, keeping zap out of call sites.
type Logger struct {
	base *zap.Logger
}

// New builds a Logger from cfg. JSON output targets stdout; console encoding is
// intended for local development.
func New(cfg Config) (log *Logger, err error) {
	levelText := cfg.Level
	if levelText == constant.EmptyString {
		levelText = constant.LogLevelInfo
	}

	encoding := cfg.Encoding
	if encoding == constant.EmptyString {
		encoding = constant.LogEncodingJSON
	}
	if encoding != constant.LogEncodingJSON && encoding != constant.LogEncodingConsole {
		err = fmt.Errorf("logger: unsupported encoding %q", encoding)
		return
	}

	var level zapcore.Level
	if level, err = zapcore.ParseLevel(strings.ToLower(levelText)); err != nil {
		err = fmt.Errorf("logger: invalid level %q: %w", levelText, err)
		return
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.Encoding = encoding
	zapConfig.OutputPaths = []string{constant.LogOutputStdout}
	zapConfig.ErrorOutputPaths = []string{constant.LogOutputStderr}
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var base *zap.Logger
	if base, err = zapConfig.Build(); err != nil {
		err = fmt.Errorf("logger: build: %w", err)
		return
	}

	log = &Logger{base: base}
	return
}

// baseLogger returns the underlying zap logger, falling back to a no-op logger
// so calls on a nil or uninitialised Logger never panic.
func (l *Logger) baseLogger() (base *zap.Logger) {
	if l != nil && l.base != nil {
		base = l.base
		return
	}
	base = zap.NewNop()
	return
}

// newWithCore builds a Logger around an existing core. It exists so tests can
// observe emitted entries without writing to stdout.
func newWithCore(core zapcore.Core) (l *Logger) {
	l = &Logger{base: zap.New(core)}
	return
}
