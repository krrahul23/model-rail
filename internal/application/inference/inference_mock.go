package inference

import (
	"context"

	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// MockInferenceService is a configurable test double for InferenceService.
type MockInferenceService struct {
	Result      response.Result
	Err         error
	Calls       int
	LastRequest request.CanonicalRequest
}

// NewMockInferenceService builds a test double with a canned result and error.
func NewMockInferenceService(result response.Result, err error) *MockInferenceService {
	return &MockInferenceService{
		Result: result,
		Err:    err,
	}
}

// Execute records the request and returns the configured result or error.
func (m *MockInferenceService) Execute(ctx context.Context, req request.CanonicalRequest) (result response.Result, err error) {
	m.Calls++
	m.LastRequest = req
	result = m.Result
	err = m.Err
	return
}

// Compile-time assertion that the test double satisfies the service contract.
var _ InferenceService = (*MockInferenceService)(nil)
