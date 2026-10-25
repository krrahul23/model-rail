// Package inference contains the inference workflow service and the mapping of
// provider failures into the gateway taxonomy.
package inference

import (
	"context"

	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// InferenceService is the domain workflow consumed by the application layer.
type InferenceService interface {
	// Complete executes a unary inference workflow.
	Complete(ctx context.Context, req request.CanonicalRequest) (response.Result, error)
}
