// Package ports declares the interfaces (ports) that the domain consumes and
// that infrastructure implements. Keeping them here lets the domain depend on
// abstractions without importing concrete infrastructure.
package ports

import (
	"context"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// Provider is the contract every LLM provider adapter implements. Calls must
// honour context cancellation and deadlines, and failures must be reported as
// response.ProviderError rather than SDK-specific error types.
type Provider interface {
	// Name returns the stable provider identifier used in metrics and routing.
	Name() string

	// Capabilities reports what the model supports. Routing filters candidates
	// on this information before any provider call is made.
	Capabilities(model constant.ModelRef) response.Capabilities

	// Invoke performs a unary completion and returns the normalized result.
	// The returned error, when non-nil, is or wraps a response.ProviderError.
	Invoke(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (response.Result, error)

	// OpenStream opens a streaming completion. The caller owns the returned
	// stream and must Close it; the adapter owns and closes the event channel.
	OpenStream(ctx context.Context, model constant.ModelRef, req request.CanonicalRequest) (ProviderStream, error)

	// Probe performs the minimal, spend-capped health check used for half-open
	// circuit recovery. It is never used for continuous probing.
	Probe(ctx context.Context, model constant.ModelRef) (response.ProbeResult, error)
}

// ProviderStream is a single-consumer event stream. The adapter closes Events
// exactly once, after a terminal event or context cancellation. Close is
// idempotent, non-blocking and safe to call from another goroutine.
type ProviderStream interface {
	// Events returns the event channel. The consumer must never close it.
	Events() <-chan response.StreamEvent

	// Close releases the upstream connection and unblocks any pending send.
	Close() error
}

// Classifier maps arbitrary provider failures onto the shared taxonomy. It must
// not leak provider SDK types across the boundary.
type Classifier interface {
	// Classify returns the classified form of err, defaulting to ClassUnknown
	// when the failure cannot be identified.
	Classify(err error) response.ProviderError
}
