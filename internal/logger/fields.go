package logger

import (
	"time"

	"modelrail/internal/constant"

	"go.uber.org/zap"
)

// Field is zap's structured field type. Callers construct fields through the
// helpers in this package, so only this package imports zap directly.
type Field = zap.Field

// String builds a string field.
func String(key, value string) (field Field) {
	field = zap.String(key, value)
	return
}

// Int builds an int field.
func Int(key string, value int) (field Field) {
	field = zap.Int(key, value)
	return
}

// Int64 builds an int64 field.
func Int64(key string, value int64) (field Field) {
	field = zap.Int64(key, value)
	return
}

// Float64 builds a float64 field.
func Float64(key string, value float64) (field Field) {
	field = zap.Float64(key, value)
	return
}

// Bool builds a bool field.
func Bool(key string, value bool) (field Field) {
	field = zap.Bool(key, value)
	return
}

// Duration builds a duration field, encoded in the logger's duration format.
func Duration(key string, value time.Duration) (field Field) {
	field = zap.Duration(key, value)
	return
}

// Err builds the conventional "error" field.
func Err(err error) (field Field) {
	field = zap.Error(err)
	return
}

// Any builds a field from an arbitrary value using zap's reflection encoder.
func Any(key string, value any) (field Field) {
	field = zap.Any(key, value)
	return
}

// RequestID records the gateway request identifier used for correlation.
func RequestID(value string) (field Field) {
	field = zap.String(constant.FieldRequestID, value)
	return
}

// TraceID records a distributed-trace identifier when one exists. Until
// OpenTelemetry is wired, callers pass the request identifier.
func TraceID(value string) (field Field) {
	field = zap.String(constant.FieldTraceID, value)
	return
}
