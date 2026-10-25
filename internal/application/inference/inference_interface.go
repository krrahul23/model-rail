// Package inference validates inference requests and dispatches them to the
// domain workflow. It contains no orchestration; the domain owns the workflow.
package inference

import (
	"context"

	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// InferenceService is the application use case exposed to transports.
type InferenceService interface {
	// Execute validates and runs an inference request.
	Execute(ctx context.Context, req request.CanonicalRequest) (response.Result, error)
}
