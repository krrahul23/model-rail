// Package httpmiddleware contains HTTP-only middleware shared by the public
// routes. It imports no controller package, so controllers can depend on it
// without creating a cycle.
package httpmiddleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/logger"
)

// requestIDContextKey isolates the request identifier in the request context.
type requestIDContextKey struct{}

// WithRequestContext assigns a request ID, installs a request-scoped logger and
// records the completion of every request.
func WithRequestContext(next http.Handler) (wrapped http.Handler) {
	wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		requestLogger := logger.L().With(
			logger.RequestID(requestID),
			logger.TraceID(requestID),
			logger.String(constant.FieldMethod, r.Method),
			logger.String(constant.FieldPath, r.URL.Path),
		)
		ctx := logger.NewContext(r.Context(), requestLogger)
		ctx = context.WithValue(ctx, requestIDContextKey{}, requestID)

		w.Header().Set(constant.HeaderRequestID, requestID)
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(recorder, r.WithContext(ctx))

		requestLogger.Info("request completed",
			logger.Int(constant.FieldStatus, recorder.status),
			logger.Duration(constant.FieldDuration, time.Since(start)),
		)
	})
	return
}

// RequestIDFromContext returns the request identifier stored by the middleware.
func RequestIDFromContext(ctx context.Context) (requestID string) {
	if ctx != nil {
		if value, ok := ctx.Value(requestIDContextKey{}).(string); ok {
			requestID = value
		}
	}
	return
}

// NewContext returns a child context carrying the request identifier. It exists
// so tests can exercise handlers without the full middleware chain.
func NewContext(ctx context.Context, requestID string) (derived context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	derived = context.WithValue(ctx, requestIDContextKey{}, requestID)
	return
}

// newRequestID returns a random hexadecimal identifier, falling back to a
// timestamp when the system entropy source is unavailable.
func newRequestID() (requestID string) {
	buffer := make([]byte, constant.RequestIDBytes)
	if _, err := rand.Read(buffer); err != nil {
		requestID = strconv.FormatInt(time.Now().UnixNano(), constant.RequestIDFallbackBase)
		return
	}
	requestID = hex.EncodeToString(buffer)
	return
}

// statusRecorder captures the status code written by a handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status before forwarding it.
func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Write records an implicit 200 status when the handler never called WriteHeader.
func (r *statusRecorder) Write(body []byte) (written int, err error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	written, err = r.ResponseWriter.Write(body)
	return
}
