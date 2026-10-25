package inference_test

import (
	"errors"
	"testing"

	applicationinference "modelrail/internal/application/inference"
	"modelrail/internal/constant"
	domainmock "modelrail/internal/domain/inference"
	"modelrail/internal/dto/request"
	"modelrail/internal/dto/response"
)

// validRequest returns a minimal request accepted by validation.
func validRequest() (req request.CanonicalRequest) {
	req = request.CanonicalRequest{
		Model: "mock-model",
		Messages: []request.Message{{
			Role:  constant.RoleUser,
			Parts: []request.ContentPart{{Type: constant.ContentText, Text: "hello"}},
		}},
	}
	return
}

func TestExecuteAppliesDefaults(t *testing.T) {
	t.Parallel()

	executor := &domainmock.MockInferenceService{}
	service := applicationinference.NewInferenceService(executor)

	if _, err := service.Execute(t.Context(), validRequest()); err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}
	if executor.Calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.Calls)
	}
	if executor.LastRequest.TaskClass != constant.TaskGeneral {
		t.Fatalf("TaskClass = %q, want %q", executor.LastRequest.TaskClass, constant.TaskGeneral)
	}
	if executor.LastRequest.MaxOutputTokens != constant.DefaultMaxOutputTokens {
		t.Fatalf("MaxOutputTokens = %d, want %d", executor.LastRequest.MaxOutputTokens, constant.DefaultMaxOutputTokens)
	}
	if executor.LastRequest.TotalDeadline != constant.DefaultRequestTimeout {
		t.Fatalf("TotalDeadline = %s, want %s", executor.LastRequest.TotalDeadline, constant.DefaultRequestTimeout)
	}
}

func TestExecuteRejectsInvalidRequests(t *testing.T) {
	temperature := constant.MaxTemperature + 1

	cases := []struct {
		name   string
		mutate func(req *request.CanonicalRequest)
	}{
		{name: "missing model", mutate: func(req *request.CanonicalRequest) {
			req.Model = constant.EmptyString
		}},
		{name: "streaming", mutate: func(req *request.CanonicalRequest) {
			req.Stream = true
		}},
		{name: "no messages", mutate: func(req *request.CanonicalRequest) {
			req.Messages = nil
		}},
		{name: "negative max tokens", mutate: func(req *request.CanonicalRequest) {
			req.MaxOutputTokens = -1
		}},
		{name: "unknown role", mutate: func(req *request.CanonicalRequest) {
			req.Messages[0].Role = "robot"
		}},
		{name: "unsupported content type", mutate: func(req *request.CanonicalRequest) {
			req.Messages[0].Parts[0].Type = "image"
		}},
		{name: "empty text", mutate: func(req *request.CanonicalRequest) {
			req.Messages[0].Parts[0].Text = constant.EmptyString
		}},
		{name: "unknown task class", mutate: func(req *request.CanonicalRequest) {
			req.TaskClass = "unknown"
		}},
		{name: "temperature out of range", mutate: func(req *request.CanonicalRequest) {
			req.Temperature = &temperature
		}},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			executor := &domainmock.MockInferenceService{}
			service := applicationinference.NewInferenceService(executor)
			req := validRequest()
			testCase.mutate(&req)

			_, err := service.Execute(t.Context(), req)
			if err == nil {
				t.Fatal("Execute() = nil, want validation error")
			}

			var gatewayErr response.GatewayError
			if !errors.As(err, &gatewayErr) {
				t.Fatalf("error = %v, want response.GatewayError", err)
			}
			if gatewayErr.Code != constant.CodeRequestInvalid {
				t.Fatalf("Code = %q, want %q", gatewayErr.Code, constant.CodeRequestInvalid)
			}
			if executor.Calls != 0 {
				t.Fatal("domain must not be called for an invalid request")
			}
		})
	}
}

func TestExecuteReturnsDomainError(t *testing.T) {
	t.Parallel()

	expected := response.GatewayError{
		Class: constant.ClassTerminalRequest,
		Code:  constant.CodeRequestInvalid,
		Cause: errors.New("domain rejected the request"),
	}
	executor := &domainmock.MockInferenceService{Err: expected}
	service := applicationinference.NewInferenceService(executor)

	_, err := service.Execute(t.Context(), validRequest())

	var gatewayErr response.GatewayError
	if !errors.As(err, &gatewayErr) {
		t.Fatalf("error = %v, want response.GatewayError", err)
	}
	if gatewayErr.Code != expected.Code {
		t.Fatalf("Code = %q, want %q", gatewayErr.Code, expected.Code)
	}
}
