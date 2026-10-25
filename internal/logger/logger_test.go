package logger

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// testLogger returns a logger backed by an observable core.
func testLogger(level zapcore.Level) (log *Logger, observed *observer.ObservedLogs) {
	core, observed := observer.New(level)
	log = newWithCore(core)
	return
}

func TestNewRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{Level: "verbose"}); err == nil {
		t.Fatal("New() = nil error, want invalid level error")
	}
}

func TestNewRejectsInvalidEncoding(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{Level: "info", Encoding: "xml"}); err == nil {
		t.Fatal("New() = nil error, want invalid encoding error")
	}
}

func TestNewAppliesConfiguredLevel(t *testing.T) {
	t.Parallel()

	log, err := New(Config{Level: "warn", Encoding: "json"})
	if err != nil {
		t.Fatalf("New() = %v, want nil", err)
	}

	core := log.baseLogger().Core()
	if core.Enabled(zapcore.DebugLevel) {
		t.Fatal("debug must be disabled at warn level")
	}
	if !core.Enabled(zapcore.WarnLevel) {
		t.Fatal("warn must be enabled at warn level")
	}
}

func TestWithAddsFields(t *testing.T) {
	t.Parallel()

	log, observed := testLogger(zapcore.DebugLevel)
	log.With(String("tenant_id", "t1")).Info("request", Int("attempt", 2))

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	fields := entries[0].ContextMap()
	if fields["tenant_id"] != "t1" {
		t.Fatalf("tenant_id = %v, want %q", fields["tenant_id"], "t1")
	}
	if fields["attempt"] != int64(2) {
		t.Fatalf("attempt = %v, want 2", fields["attempt"])
	}
}

func TestFieldHelpers(t *testing.T) {
	t.Parallel()

	log, observed := testLogger(zapcore.DebugLevel)
	log.Info("fields",
		String("string", "value"),
		Int("int", 1),
		Int64("int64", 2),
		Float64("float64", 1.5),
		Bool("bool", true),
		Duration("duration", time.Second),
		Err(errors.New("boom")),
		Any("any", map[string]string{"k": "v"}),
		RequestID("req-1"),
		TraceID("trace-1"),
	)

	fields := observed.All()[0].ContextMap()
	for _, key := range []string{
		"string", "int", "int64", "float64", "bool",
		"duration", "error", "any", "request_id", "trace_id",
	} {
		if _, ok := fields[key]; !ok {
			t.Fatalf("field %q is missing from the emitted entry", key)
		}
	}
}

func TestNilLoggerIsSafe(t *testing.T) {
	t.Parallel()

	var log *Logger
	log.Info("no panic")
	log.With(String("component", "test")).Warn("no panic")
}

func TestGlobalLogger(t *testing.T) {
	// Not parallel: this test mutates the process-wide logger.
	log, observed := testLogger(zapcore.DebugLevel)
	SetGlobal(log)
	t.Cleanup(func() {
		SetGlobal(nil)
	})

	Info("global", String("component", "test"))

	entries := observed.All()
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	if entries[0].Message != "global" {
		t.Fatalf("message = %q, want %q", entries[0].Message, "global")
	}
}

func TestContextLogger(t *testing.T) {
	t.Parallel()

	log, observed := testLogger(zapcore.DebugLevel)
	ctx := NewContext(context.Background(), log)

	FromContext(ctx).Info("from context")

	if got := len(observed.All()); got != 1 {
		t.Fatalf("got %d entries, want 1", got)
	}
}

func TestFromContextFallsBackToGlobal(t *testing.T) {
	// Not parallel: depends on the process-wide logger.
	log, observed := testLogger(zapcore.DebugLevel)
	SetGlobal(log)
	t.Cleanup(func() {
		SetGlobal(nil)
	})

	FromContext(context.Background()).Info("fallback")

	if got := len(observed.All()); got != 1 {
		t.Fatalf("got %d entries, want 1", got)
	}
}
