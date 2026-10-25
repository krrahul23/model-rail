package inference

import (
	"context"
	"errors"

	"modelrail/internal/constant"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
	"modelrail/internal/ports"
)

// inferenceService is the concrete workflow implementation.
type inferenceService struct {
	provider ports.Provider
}

// NewInferenceService builds the workflow around a provider adapter.
func NewInferenceService(provider ports.Provider) (service InferenceService) {
	service = &inferenceService{provider: provider}
	return
}

// Complete executes a unary completion under the request deadline and maps any
// provider failure into the gateway error taxonomy.
func (s *inferenceService) Complete(ctx context.Context, req request.CanonicalRequest) (result response.Result, err error) {
	callCtx := ctx
	if req.TotalDeadline > 0 {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, req.TotalDeadline)
		defer cancel()
	}

	result, err = s.provider.Invoke(callCtx, req.Model, req)
	if err != nil {
		err = mapProviderError(err)
		return
	}
	return
}

// mapProviderError converts an adapter failure into the gateway taxonomy. The
// provider's cause is preserved for diagnostics but never exposed verbatim to
// clients.
func mapProviderError(err error) (gatewayErr response.GatewayError) {
	var providerErr response.ProviderError
	if errors.As(err, &providerErr) {
		gatewayErr = response.GatewayError{
			Class:      providerErr.Class,
			Code:       codeForProviderError(providerErr),
			RetryAfter: providerErr.RetryAfter,
			Cause:      err,
		}
		return
	}

	// An unclassified failure is treated as an upstream outage.
	gatewayErr = response.GatewayError{
		Class: constant.ClassUnknown,
		Code:  constant.CodeProviderUnavailable,
		Cause: err,
	}
	return
}

// codeForProviderError selects the stable client-visible code for a provider
// failure class.
func codeForProviderError(err response.ProviderError) (code constant.ErrorCode) {
	switch err.Class {
	case constant.ClassThrottled:
		code = constant.CodeProviderThrottled
	case constant.ClassTransient:
		// A deadline is reported as a timeout; other transient failures are
		// reported as an unavailable provider.
		code = constant.CodeProviderUnavailable
		if err.ProviderCode == constant.ProviderCodeContextDeadlineExceeded {
			code = constant.CodeProviderTimeout
		}
	case constant.ClassPlatformConfig:
		code = constant.CodePlatformConfigError
	case constant.ClassTerminalRequest:
		code = constant.CodeRequestInvalid
	case constant.ClassDegradedStream:
		code = constant.CodeDegradedStream
	default:
		code = constant.CodeProviderUnavailable
	}
	return
}
