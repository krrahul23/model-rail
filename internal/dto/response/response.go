// Package response defines the canonical response DTOs and error values
// produced by the pipeline and consumed by the controller.
package response

import (
	"time"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
)

// Usage is the token accounting for one request. Source records whether the
// numbers came from the provider or from the local estimator.
type Usage struct {
	InputTokens       int64
	CachedInputTokens int64
	OutputTokens      int64
	ReasoningTokens   int64
	Source            constant.UsageSource
}

// LatencyBreakdown separates time to first token from total time, so gateway
// overhead and provider latency can be reported independently.
type LatencyBreakdown struct {
	TTFT  time.Duration
	Total time.Duration
}

// Result is the normalized outcome of a unary completion. ModelServed always
// reports the model that actually served the request; it can differ from the
// requested model when an opted-in fallback group is used.
type Result struct {
	Content      []request.ContentPart
	Usage        Usage
	ModelServed  constant.ModelRef
	ProviderName string
	FinishReason constant.FinishReason
	Latency      LatencyBreakdown
}

// Capabilities describes what a model supports. Routing filters candidates on
// these fields before any provider call is made.
type Capabilities struct {
	ContextWindow           int
	SupportsStreaming       bool
	SupportsTools           bool
	SupportsJSONMode        bool
	SupportsCachedTokens    bool
	SupportsReasoningTokens bool
	TieredPricing           bool
}

// ProbeResult is the minimal health-check outcome used for half-open circuit
// recovery. It is not used for continuous probing.
type ProbeResult struct {
	Healthy bool
	Latency time.Duration
}
