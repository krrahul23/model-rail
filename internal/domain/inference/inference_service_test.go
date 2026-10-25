package inference_test

import (
	"errors"
	"testing"
	"time"

	"modelrail/internal/constant"
	domaininference "modelrail/internal/domain/inference"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
	"modelrail/internal/ports"
)

// durationPtr returns a pointer to d for optional fields.
func durationPtr(d time.Duration) (ptr *time.Duration) {
	ptr = &d
	return
}

// validRequest returns a minimal canonical request.
func validRequest() (req request.CanonicalRequest) {
	req = request.CanonicalRequest{Model: "mock-model"}
	return
}

func TestCompleteReturnsProviderResult(t *testing.T) {
	t.Parallel()

	provider := &ports.MockProvider{InvokeResult: response.Result{
		Content:      []request.ContentPart{{Type: constant.ContentText, Text: "hello"}},
		Usage:        response.Usage{InputTokens: 1, OutputTokens: 2, Source: constant.UsageReported},
		ModelServed:  "mock-model",
		ProviderName: constant.ProviderMock,
		FinishReason: constant.FinishStop,
	}}
	service := domaininference.NewInferenceService(provider)

	result, err := service.Complete(t.Context(), validRequest())
	if err != nil {
		t.Fatalf("Complete() = %v, want nil", err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "hello" {
		t.Fatalf("Content = %+v, want the configured response", result.Content)
	}
	if result.ModelServed != "mock-model" {
		t.Fatalf("ModelServed = %q, want %q", result.ModelServed, "mock-model")
	}
	if result.ProviderName != constant.ProviderMock {
		t.Fatalf("ProviderName = %q, want %q", result.ProviderName, constant.ProviderMock)
	}
	if provider.InvokeCalls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.InvokeCalls)
	}
	if provider.LastModel != "mock-model" {
		t.Fatalf("LastModel = %q, want %q", provider.LastModel, "mock-model")
	}
}

func TestCompleteMapsThrottledError(t *testing.T) {
	t.Parallel()

	retryAfter := 2 * time.Second
	providerErr := response.ProviderError{
		Class:        constant.ClassThrottled,
		Provider:     constant.ProviderMock,
		ProviderCode: "429",
		RetryAfter:   durationPtr(retryAfter),
	}
	service := domaininference.NewInferenceService(&ports.MockProvider{InvokeErr: providerErr})

	_, err := service.Complete(t.Context(), validRequest())

	var gatewayErr response.GatewayError
	if !errors.As(err, &gatewayErr) {
		t.Fatalf("error = %v, want response.GatewayError", err)
	}
	if gatewayErr.Code != constant.CodeProviderThrottled {
		t.Fatalf("Code = %q, want %q", gatewayErr.Code, constant.CodeProviderThrottled)
	}
	if gatewayErr.RetryAfter == nil || *gatewayErr.RetryAfter != retryAfter {
		t.Fatalf("RetryAfter = %v, want %s", gatewayErr.RetryAfter, retryAfter)
	}
}

func TestCompleteMapsPlatformConfigError(t *testing.T) {
	t.Parallel()

	providerErr := response.ProviderError{
		Class:        constant.ClassPlatformConfig,
		Provider:     constant.ProviderMock,
		ProviderCode: "invalid_key",
	}
	service := domaininference.NewInferenceService(&ports.MockProvider{InvokeErr: providerErr})

	_, err := service.Complete(t.Context(), validRequest())

	var gatewayErr response.GatewayError
	if !errors.As(err, &gatewayErr) {
		t.Fatalf("error = %v, want response.GatewayError", err)
	}
	if gatewayErr.Code != constant.CodePlatformConfigError {
		t.Fatalf("Code = %q, want %q", gatewayErr.Code, constant.CodePlatformConfigError)
	}
}

func TestCompleteAppliesDeadlineAndMapsTimeout(t *testing.T) {
	t.Parallel()

	provider := &ports.MockProvider{InvokeDelay: 100 * time.Millisecond}
	service := domaininference.NewInferenceService(provider)
	req := validRequest()
	req.TotalDeadline = 5 * time.Millisecond

	_, err := service.Complete(t.Context(), req)

	var gatewayErr response.GatewayError
	if !errors.As(err, &gatewayErr) {
		t.Fatalf("error = %v, want response.GatewayError", err)
	}
	if gatewayErr.Code != constant.CodeProviderTimeout {
		t.Fatalf("Code = %q, want %q", gatewayErr.Code, constant.CodeProviderTimeout)
	}
	if provider.InvokeCalls != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.InvokeCalls)
	}
}
